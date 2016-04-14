package main

import (
	"control"
	"encoding/gob"
	//"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
	"os/exec"
	"reflect"
)

const IP1 = "129.241.187.147" //Start router on this IP
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
var elevatorWhichSentTheOrder string

var elevatorEncoders = make(map[string]*gob.Encoder)
var elevatorDecoders = make(map[string]*gob.Decoder)

var matrixInTransit = make(map[string]control.ElevatorNode)

var sendMatrix bool

var backupAliveConnection net.Conn
var backupCommConnection net.Conn
var backupEncoder *gob.Encoder

var backupIsDead bool

var elevatorWhichSentTheOrderMutex = &sync.Mutex{}
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
	backupAliveConnection, _ = backupListener.Accept()
	backupCommConnection, _ = backupListener.Accept()
	getBackupIP()
	fmt.Println("Connected to backup")
	backupEncoder = gob.NewEncoder(backupCommConnection)
}


func receiveIncoming(dec *gob.Decoder, channel chan map[string]control.ElevatorNode){
	var newMap = make(map[string]control.ElevatorNode)
	for{
		dec.Decode(&newMap)
		//fmt.Println(newMap)
		channel <- newMap
	}
}


func connectNewElevatorsThread(wg *sync.WaitGroup, channel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		aliveConnection, err := elevatorListener.Accept()
		if err != nil {
			panic(err)
		}
		commConnection, err := elevatorListener.Accept()
		if err != nil {
			panic(err)
		}
		elevatorIPAddress := aliveConnection.RemoteAddr().String()
		elevatorAliveConnections[elevatorIPAddress] = aliveConnection
		elevatorCommConnections[elevatorIPAddress] = commConnection

		elevatorEncoders[elevatorIPAddress] = gob.NewEncoder(commConnection)
		elevatorDecoders[elevatorIPAddress] = gob.NewDecoder(commConnection)

		go receiveIncoming(elevatorDecoders[elevatorIPAddress], channel)
		wg.Add(1)


		var tempMatrix = make(map[string]control.ElevatorNode)
		elevatorDecoders[elevatorIPAddress].Decode(&tempMatrix)
		connectionMutex.Lock()
		initialNode := tempMatrix[elevatorIPAddress]
		matrixInTransit[elevatorIPAddress] = initialNode
		connectionMutex.Unlock()
		for elevator, _ := range elevatorAliveConnections {
			elevatorEncoders[elevator].Encode(matrixInTransit)
		}
	}
}


func tellElevatorStillConnected(elevatorIP string) bool{
	text := "Still alive"
	_, err := fmt.Fprintf(elevatorAliveConnections[elevatorIP], text)
	if err != nil{
		return false
	}
	return true
}



func tellElevatorStillConnectedThread(){
	for{
		time.Sleep(time.Millisecond * 500)
		for elevator, _ := range elevatorAliveConnections{
			if !tellElevatorStillConnected(elevator){
				elevatorIsDead(elevator)
			}
		}
	}
}

//Other errors than cut network connection could kill elevator
func checkElevatorStillConnected(elevatorIP string) bool {
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
		time.Sleep(time.Millisecond * 500)
		for elevator, _ := range elevatorAliveConnections {
			if !checkElevatorStillConnected(elevator) {
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
		//fmt.Println("Sending new map without dead elevator to")
		//fmt.Println(elevator)

	}
	//fmt.Println("Elevator died. New map")
	//fmt.Println(elevatorCommConnections)
	//fmt.Println("This is the map without the dead elevator")
	//fmt.Println(matrixInTransit)

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


func getMatrixThread(channel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		tempMatrix := <- channel
		if !reflect.DeepEqual(matrixInTransit, tempMatrix) {
			connectionMutex.Lock()
			copyMapByValue(tempMatrix, matrixInTransit)
			fmt.Println("This will be sent onwards")
			fmt.Println(tempMatrix)
			connectionMutex.Unlock()
			sendMatrix = true
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
				elevatorEncoders[elevator].Encode(tempMatrix)
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
	elevatorChannel := make(chan map[string]control.ElevatorNode)

	wg := new(sync.WaitGroup)
	wg.Add(7)
	routerModuleInit()
	go connectNewElevatorsThread(wg, elevatorChannel)
	go checkElevatorStillConnectedThread()
	go tellElevatorStillConnectedThread()
	go tellBackupAliveThread()
	go spawnNewBackupThread()
	go getMatrixThread(elevatorChannel)
	go sendMatrixThread()

	wg.Wait()
}
