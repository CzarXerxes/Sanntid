package main

import (
	"control"
	//"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"
	//"bufio"
	"os/exec"
)

var backupListener net.Listener
var elevatorListener net.Listener

var elevatorTracking = make(map[string]bool)
var elevatorConnections = make(map[string]net.Conn) //Dictionary with ipAddress:connectionSocket

var matrixInTransit map[string]control.ElevatorNode
var sendMatrix bool

var backupIPAddress = "129.241.187.153"

var backupConnection net.Conn
var backupIsDead bool

//var backupEncoder gob.Encoder

var connectionMutex = &sync.Mutex{}

const backupPort = ":30000"
const elevatorPort = ":29000"

func routerModuleInit() {
	backupListener, _ = net.Listen("tcp", backupPort)
	elevatorListener, _ = net.Listen("tcp", elevatorPort)
	backupConnection = spawnBackup()
}

func spawnBackup() net.Conn {
	var tempConnection net.Conn
	fmt.Println("Making a backup")
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run BackupModule.go")
	_ = cmd.Run()
	for {
		tempConnection, _ = backupListener.Accept()
		fmt.Println("Connected to backup")
		break
	}
	return tempConnection
}

//Implement this to send elevators to backup
func sendElevatorMapToBackup() {
	enc := gob.NewEncoder(backupConnection)
	enc.Encode(elevatorConnections)
}

//Fix this bug: Adding new conn to map overwrites previous conn
func assignElevatorAddress(conn net.Conn) {
	elevatorConnections[conn.RemoteAddr().String()] = conn
}

func addNewElevatorsToTracking(conn net.Conn) {
	elevatorTracking[conn.RemoteAddr().String()] = true
}

func connectNewElevatorsThread() {
	//addr, _ := net.ResolveTCPAddr("tcp", port)
	//ln, _ := net.Listen("tcp", port)
	for {
		time.Sleep(time.Millisecond * 100)
		//ln, _ := net.Listen("tcp", ":30000")
		//var connection = *new(net.Conn)
		fmt.Println(elevatorConnections)
		fmt.Println(elevatorTracking)
		connection, _ := elevatorListener.Accept()
		connectionMutex.Lock()
		fmt.Println("Found elevator")
		//time.Sleep(time.Second * 1)
		assignElevatorAddress(connection)
		addNewElevatorsToTracking(connection)
		sendElevatorMapToBackup()
		connectionMutex.Unlock()
	}
}

//Other errors than cut network connection could kill elevator
func elevatorStillConnected(elevatorIP string) bool {
	socket := elevatorConnections[elevatorIP]
	buf := make([]byte, 1024)
	_, err := socket.Read(buf)
	if err != nil {
		return false
	}
	//fmt.Printf("Message received :: %s\n", string(buf[:n]))
	return true
}

func checkElevatorStillConnectedThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		for elevator, _ := range elevatorConnections {
			if !elevatorStillConnected(elevator) {
				elevatorIsDead(elevator)
			}
		}
	}
}

func elevatorIsDead(elevator string) {
	delete(elevatorTracking, elevator)
	elevatorConnections[elevator].Close()
	delete(elevatorConnections, elevator)
	fmt.Println(elevatorConnections)
}

func tellBackupAliveThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		text := "Router is still alive"
		fmt.Fprintf(backupConnection, text)
	}
}

func backupIsAlive() bool {
	buf := make([]byte, 1024)
	_, err := backupConnection.Read(buf)
	if err != nil {
		return false
	}
	return true
	//fmt.Println("Receiving im alive")
}

func spawnNewBackupThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		if !backupIsAlive() {
			backupIPAddress = "129.241.187.153"
			//fmt.Println("Locking mutex")
			connectionMutex.Lock()
			//fmt.Println("Mutex locked")
			backupConnection.Close()
			connectionMutex.Unlock()
			//time.Sleep(time.Second * 1)
			connectionMutex.Lock()
			backupConnection = spawnBackup()
			connectionMutex.Unlock()
			//fmt.Println("Mutex unlocked")
		}
	}
}



func getMatrixThread() {
	for {
		time.Sleep(time.Millisecond * 10)
		for elevator, _ := range elevatorConnections {
			connectionMutex.Lock()
			dec := gob.NewDecoder(elevatorConnections[elevator])
			deadline := time.Now().Add(time.Millisecond * 1)
			for time.Now().Before(deadline) {
				dec.Decode(&matrixInTransit)
			}
			//fmt.Println("%#v", matrixInTransit)
			sendMatrix = true
		}
	}
}

func sendMatrixThread() {
	for {
		time.Sleep(time.Millisecond * 10)
		if sendMatrix {
			for elevator, _ := range elevatorConnections {
				for i := 0; i < 10; i++ {
					//time.Sleep(time.Millisecond * 10)
					enc := gob.NewEncoder(elevatorConnections[elevator])
					fmt.Println("Sending this back to elevator")
					fmt.Println("%#v", matrixInTransit)
					enc.Encode(matrixInTransit)
				}
				/*
					enc := gob.NewEncoder(elevatorConnections[elevator])
					fmt.Println("Sending this back to elevator")
					fmt.Println("%#v", matrixInTransit)
					enc.Encode(matrixInTransit)
				*/
			}
			sendMatrix = false
			connectionMutex.Unlock()
		}
	}
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(6)
	routerModuleInit()
	go connectNewElevatorsThread()
	go checkElevatorStillConnectedThread()
	go tellBackupAliveThread()
	go spawnNewBackupThread()
	//go transferMatrixThread()
	go getMatrixThread()
	go sendMatrixThread()

	wg.Wait()
}
