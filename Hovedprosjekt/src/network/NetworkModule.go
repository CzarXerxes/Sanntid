package network

import (
	//"bufio"
	"control"
	"encoding/gob"
	"fmt"
	"net"
	//"os"
	//"runtime"
	"reflect"
	"sync"
	"time"
)

const IP = "129.241.187.144"

//Changes with workspace
var routerIPAddress = IP

//const port = "20021"
const routerPort = "29000"
const backupPort = "28000"

//General connection variables
//var tcpSendConnection net.Conn
var routerAliveConnection net.Conn
var routerCommConnection net.Conn

var routerEncoder *gob.Encoder
var routerDecoder *gob.Decoder

var elevatorMatrixMutex = &sync.Mutex{}
var sendMatrixToRouter bool
var sendMatrixToElevator bool
var routerIsDead bool

var matrixInTransit map[string]control.ElevatorNode
var matrixMostRecentlySent map[string]control.ElevatorNode

func getIPAddress() string {
	address := routerAliveConnection.LocalAddr().String()
	return address
}

func sendInitialAddressToElevator(address string, initializeAddressChannel chan string) {
	initializeAddressChannel <- address
}

func getRouterConnection() bool {
	var err error
	routerAliveConnection, err = net.Dial("tcp", net.JoinHostPort(routerIPAddress, routerPort))
	if err != nil {
		return false
	}
	routerCommConnection, err = net.Dial("tcp", net.JoinHostPort(routerIPAddress, routerPort))
	if err != nil {
		return false
	}
	routerEncoder = gob.NewEncoder(routerCommConnection)
	routerDecoder = gob.NewDecoder(routerCommConnection)
	return true
}

func networkModuleInit(firstTimeCalled bool, initializeAddressChannel chan string, blockNetworkChan chan bool, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	var tempMatrix = make(map[string]control.ElevatorNode)
	matrixInTransit = make(map[string]control.ElevatorNode)
	matrixMostRecentlySent = make(map[string]control.ElevatorNode)
	if firstTimeCalled {
		waitBecauseElevatorsHavePreviouslyCrashed := <-blockNetworkChan
		if waitBecauseElevatorsHavePreviouslyCrashed {
			waitBecauseElevatorsHavePreviouslyCrashed = <-blockNetworkChan
		}
	}

	for !getRouterConnection() {
		sendInitialAddressToElevator("0", initializeAddressChannel)
	}
	localAddress := getIPAddress()
	sendInitialAddressToElevator(localAddress, initializeAddressChannel)
	tempMatrix = <-receiveFromElevatorChannel
	sendToRouter(tempMatrix)
	time.Sleep(time.Millisecond * 500)
	tempMatrix = receiveFromRouter()
	elevatorMatrixMutex.Lock()
	copyMapByValue(tempMatrix, matrixInTransit)
	sendToElevatorChannel <- tempMatrix
	elevatorMatrixMutex.Unlock()
}

func closeNetworkConnection() {
	routerAliveConnection.Close()
	routerCommConnection.Close()
}

//Communicating with router functions

func sendToRouter(matrixInTransit map[string]control.ElevatorNode) {
	var tempData = make(map[string]control.ElevatorNode)
	copyMapByValue(matrixInTransit, tempData)
	routerEncoder.Encode(tempData)

}

func sendToRouterThread() {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if sendMatrixToRouter {
			//fmt.Println("Transferring this matrix to router")
			//fmt.Println(matrixInTransit)
			elevatorMatrixMutex.Lock()
			copyMapByValue(matrixInTransit, tempMatrix)
			copyMapByValue(matrixInTransit, matrixMostRecentlySent)
			elevatorMatrixMutex.Unlock()
			sendToRouter(tempMatrix)
			sendMatrixToRouter = false
		}
	}
}

func receiveFromRouter() map[string]control.ElevatorNode {
	var receivedData = make(map[string]control.ElevatorNode)
	var tempData = make(map[string]control.ElevatorNode)
	routerDecoder.Decode(&receivedData)
	copyMapByValue(receivedData, tempData)
	return tempData
}

func receiveFromRouterThread() {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if !sendMatrixToRouter {
			tempMatrix = receiveFromRouter()
			elevatorMatrixMutex.Lock()
			if !reflect.DeepEqual(tempMatrix, matrixMostRecentlySent) {
				copyMapByValue(tempMatrix, matrixInTransit)
				//fmt.Println("Received this matrix from router")
				//fmt.Println(tempMatrix)
			}
			sendMatrixToElevator = true
			elevatorMatrixMutex.Unlock()
		}
	}
}

func communicateWithRouterThread() {
	go sendToRouterThread()
	go receiveFromRouterThread()
}

//Thread to tell router module that this elevator is still connected to the network
func tellRouterStillAlive() bool {
	text := "Still alive"
	if routerAliveConnection == nil {
		return false
	}
	_, err := fmt.Fprintf(routerAliveConnection, text)
	if err != nil {
		return false
	}
	return true
}

func checkRouterStillAlive() bool {
	buf := make([]byte, 1024)
	if routerAliveConnection == nil {
		return false
	}
	_, err := routerAliveConnection.Read(buf)
	if err != nil {
		return false
	}
	return true
}

func tellRouterStillAliveThread(initialAddressChannel chan string, blockNetworkChan chan bool, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 500)
		if !tellRouterStillAlive() {
			routerIPAddress = nextRouterIP()
			networkModuleInit(false, initialAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)
			time.Sleep(time.Millisecond * 500)
		}
	}

}

func checkRouterStillAliveThread(initialAddressChannel chan string, blockNetworkChan chan bool, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 500)
		if !checkRouterStillAlive() {
			routerIPAddress = nextRouterIP()
			networkModuleInit(false, initialAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func nextRouterIP() string {
	return IP
}

//Communication with elevator functions
func communicateWithElevatorThread(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	go receiveFromElevatorThread(receiveChannel)
	go sendToElevatorThread(sendChannel)
}

func receiveFromElevatorThread(receiveChannel chan map[string]control.ElevatorNode) {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if !sendMatrixToElevator {
			tempMatrix = <-receiveChannel
			if !reflect.DeepEqual(matrixInTransit, tempMatrix) {
				elevatorMatrixMutex.Lock()
				copyMapByValue(tempMatrix, matrixInTransit)
				elevatorMatrixMutex.Unlock()
				sendMatrixToRouter = true
			}
		}
	}
}

func sendToElevatorThread(sendChannel chan map[string]control.ElevatorNode) {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		//fmt.Println(matrixInTransit)
		time.Sleep(time.Millisecond * 10)
		if sendMatrixToElevator {
			elevatorMatrixMutex.Lock()
			copyMapByValue(matrixInTransit, tempMatrix)
			elevatorMatrixMutex.Unlock()
			sendChannel <- tempMatrix
			sendMatrixToElevator = false
		}
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

func Run(initializeAddressChannel chan string, blockNetworkChan chan bool, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(3)
	networkModuleInit(true, initializeAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)

	go communicateWithElevatorThread(sendToElevatorChannel, receiveFromElevatorChannel)
	go communicateWithRouterThread()
	go checkRouterStillAliveThread(initializeAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)
	go tellRouterStillAliveThread(initializeAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)

	wg.Wait()
	closeNetworkConnection()
}
