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

var elevatorAliveConnections = make(map[string]net.Conn) //Dictionary with ipAddress:connectionSocket
var elevatorCommConnections = make(map[string]net.Conn)

var elevatorEncoders = make(map[string]*gob.Encoder)
var elevatorDecoders = make(map[string]*gob.Decoder)

var matrixInTransit = make(map[string]control.ElevatorNode)
var sendMatrix bool

var backupIPAddress = "129.241.187.153"

var backupAliveConnection net.Conn
var backupCommConnection net.Conn
var backupEncoder *gob.Encoder

var backupIsDead bool

var connectionMutex = &sync.Mutex{}

const backupPort = ":30000"
const elevatorPort = ":29000"

func routerModuleInit() {
	backupListener, _ = net.Listen("tcp", backupPort)
	elevatorListener, _ = net.Listen("tcp", elevatorPort)
	spawnBackup()
}

func spawnBackup() {
	fmt.Println("Making a backup")
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run BackupModule.go")
	_ = cmd.Run()
	//for {
	backupAliveConnection, _ = backupListener.Accept()
	backupCommConnection, _ = backupListener.Accept()
	fmt.Println("Connected to backup")
	//	break
	//}
	backupEncoder = gob.NewEncoder(backupCommConnection)
}

//Implement this to send elevators to backup
func sendElevatorMapToBackup() {
	backupEncoder.Encode(elevatorCommConnections)
}

func connectNewElevatorsThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		fmt.Println(elevatorCommConnections)
		aliveConnection, _ := elevatorListener.Accept()
		commConnection, _ := elevatorListener.Accept()
		connectionMutex.Lock()
		elevatorIPAddress := aliveConnection.RemoteAddr().String()
		elevatorAliveConnections[elevatorIPAddress] = aliveConnection
		elevatorCommConnections[elevatorIPAddress] = commConnection

		elevatorEncoders[elevatorIPAddress] = gob.NewEncoder(commConnection)
		elevatorDecoders[elevatorIPAddress] = gob.NewDecoder(commConnection)

		var tempMatrix map[string]control.ElevatorNode
		elevatorDecoders[elevatorIPAddress].Decode(&tempMatrix)
		initialNode := tempMatrix[elevatorIPAddress]
		matrixInTransit[elevatorIPAddress] = initialNode

		for elevator, _ := range elevatorAliveConnections {
			elevatorEncoders[elevator].Encode(matrixInTransit)
		}
		sendElevatorMapToBackup()
		connectionMutex.Unlock()
	}
}

//Other errors than cut network connection could kill elevator
func elevatorStillConnected(elevatorIP string) bool {
	buf := make([]byte, 1024)
	_, err := elevatorAliveConnections[elevatorIP].Read(buf)
	if err != nil {
		return false
	}
	//fmt.Printf("Message received :: %s\n", string(buf[:n]))
	return true
}

func checkElevatorStillConnectedThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		for elevator, _ := range elevatorAliveConnections {
			if !elevatorStillConnected(elevator) {
				elevatorIsDead(elevator)
			}
		}
	}
}

func elevatorIsDead(elevator string) {
	elevatorAliveConnections[elevator].Close()
	elevatorCommConnections[elevator].Close()
	time.Sleep(time.Second * 1)
	delete(elevatorAliveConnections, elevator)
	delete(elevatorCommConnections, elevator)
	delete(elevatorEncoders, elevator)
	delete(elevatorDecoders, elevator)
	delete(matrixInTransit, elevator)
	for elevator, _ := range elevatorAliveConnections {
		elevatorEncoders[elevator].Encode(matrixInTransit)

	}
	fmt.Println("Elevator died. New map")
	fmt.Println(elevatorCommConnections)
}

func tellBackupAliveThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		text := "Router is still alive"
		fmt.Fprintf(backupAliveConnection, text)
	}
}

func backupIsAlive() bool {
	buf := make([]byte, 1024)
	_, err := backupAliveConnection.Read(buf)
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
			backupAliveConnection.Close()
			backupCommConnection.Close()
			connectionMutex.Unlock()
			//time.Sleep(time.Second * 1)
			connectionMutex.Lock()
			spawnBackup()
			connectionMutex.Unlock()
			//fmt.Println("Mutex unlocked")
		}
	}
}

func getMatrixThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		for elevator, _ := range elevatorAliveConnections {
			connectionMutex.Lock()
			//deadline := time.Now().Add(time.Millisecond * 1)
			//for time.Now().Before(deadline) {
			elevatorDecoders[elevator].Decode(&matrixInTransit)
			fmt.Println("Received this from elevator")
			fmt.Println(matrixInTransit)
			//}
			sendMatrix = true
			connectionMutex.Unlock()
		}
	}
}

func sendMatrixThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		connectionMutex.Lock()
		if sendMatrix {
			for elevator, _ := range elevatorAliveConnections {
				fmt.Println("Sending this back to elevator")
				fmt.Println("%#v", matrixInTransit)
				elevatorEncoders[elevator].Encode(matrixInTransit)

			}
		}
		sendMatrix = false
		connectionMutex.Unlock()
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
