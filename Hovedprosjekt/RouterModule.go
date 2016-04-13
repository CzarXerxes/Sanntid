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
	"reflect"
)

const IP1 = "129.241.187.148" //Start router on this IP
const IP2 = "129.241.187.142"
const IP3 = "129.241.187.142"

const backupPort = ":30000"
const elevatorPort = ":29000"

var routerIPAddress string
var backupIPAddress string

var backupListener net.Listener
var elevatorListener net.Listener

var elevatorAliveConnections = make(map[string]net.Conn) //Dictionary with ipAddress:connectionSocket
var elevatorCommConnections = make(map[string]net.Conn)

var elevatorEncoders = make(map[string]*gob.Encoder)
var elevatorDecoders = make(map[string]*gob.Decoder)

var matrixInTransit = make(map[string]control.ElevatorNode)
var sendMatrix bool

var backupAliveConnection net.Conn
var backupCommConnection net.Conn
var backupEncoder *gob.Encoder

var backupIsDead bool

var connectionMutex = &sync.Mutex{}

func getRouterIP() { //Implement to find local IP address
	routerIPAddress = IP1
}

func routerModuleInit() {
	getRouterIP()
	backupListener, _ = net.Listen("tcp", backupPort)
	elevatorListener, _ = net.Listen("tcp", elevatorPort)
	spawnBackup()
}

func getBackupIP() {
	if routerIPAddress == IP1 {
		backupIPAddress = IP2
	} else if routerIPAddress == IP2 {
		backupIPAddress = IP3
	} else if routerIPAddress == IP3 {
		backupIPAddress = IP1
	}
}

func spawnBackup() {
	fmt.Println("Making a backup")
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run BackupModule.go")
	_ = cmd.Run()
	//for {
	backupAliveConnection, _ = backupListener.Accept()
	backupCommConnection, _ = backupListener.Accept()
	getBackupIP()
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
		time.Sleep(time.Millisecond * 10)
		fmt.Println(elevatorCommConnections)
		aliveConnection, _ := elevatorListener.Accept()
		commConnection, _ := elevatorListener.Accept()
		elevatorIPAddress := aliveConnection.RemoteAddr().String()
		elevatorAliveConnections[elevatorIPAddress] = aliveConnection
		elevatorCommConnections[elevatorIPAddress] = commConnection

		elevatorEncoders[elevatorIPAddress] = gob.NewEncoder(commConnection)
		elevatorDecoders[elevatorIPAddress] = gob.NewDecoder(commConnection)

		var tempMatrix = make(map[string]control.ElevatorNode)
		elevatorDecoders[elevatorIPAddress].Decode(&tempMatrix)
		connectionMutex.Lock()
		initialNode := tempMatrix[elevatorIPAddress]
		matrixInTransit[elevatorIPAddress] = initialNode
		connectionMutex.Unlock()
		for elevator, _ := range elevatorAliveConnections {
			elevatorEncoders[elevator].Encode(matrixInTransit)
		}
		sendElevatorMapToBackup()
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
		time.Sleep(time.Millisecond * 10)
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
		time.Sleep(time.Millisecond * 10)
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
		time.Sleep(time.Millisecond * 10)
		if !backupIsAlive() {
			getBackupIP()
			backupAliveConnection.Close()
			backupCommConnection.Close()
			spawnBackup()
		}
	}
}

func getMatrixThread() {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		for elevator, _ := range elevatorAliveConnections {
			deadline := time.Now().Add(time.Millisecond * 1)
			for time.Now().Before(deadline) {
				elevatorDecoders[elevator].Decode(&tempMatrix)
			}
			if !reflect.DeepEqual(matrixInTransit, tempMatrix) {
				connectionMutex.Lock()
				copyMapByValue(tempMatrix, matrixInTransit)
				connectionMutex.Unlock()
				sendMatrix = true
			}
		}
	}
}

func sendMatrixThread() {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		connectionMutex.Lock()
		copyMapByValue(matrixInTransit, tempMatrix)
		connectionMutex.Unlock()
		if sendMatrix {
			for elevator, _ := range elevatorAliveConnections {
				fmt.Println("Transmitting this across the network")
				elevatorEncoders[elevator].Encode(tempMatrix)
				fmt.Println(tempMatrix)
			}
		}
		sendMatrix = false

	}
}

func copyMapByValue(originalMap map[string]control.ElevatorNode, newMap map[string]control.ElevatorNode) {
	for k, _ := range newMap {
		delete(newMap, k)
	}
	for k, v := range originalMap {
		newMap[k] = v
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
	go getMatrixThread()
	go sendMatrixThread()

	wg.Wait()
}
