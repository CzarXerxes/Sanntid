package network

import (
	//"bufio"
	"control"
	"encoding/gob"
	"fmt"
	"net"
	//"os"
	//"runtime"
	"sync"
	"time"
)

//General connection constant

//Changes with workspace
var routerIPAddress = "129.241.187.153"

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

func networkModuleInit(initializeAddressChannel chan string, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	for !getRouterConnection() {
		sendInitialAddressToElevator("0", initializeAddressChannel)
	}
	localAddress := getIPAddress()
	sendInitialAddressToElevator(localAddress, initializeAddressChannel)
	tempMatrix := <-receiveFromElevatorChannel
	sendToRouter(tempMatrix)
	time.Sleep(time.Millisecond * 500)
	tempMatrix = receiveFromRouter()
	elevatorMatrixMutex.Lock()
	matrixInTransit = tempMatrix
	sendToElevatorChannel <- matrixInTransit
	elevatorMatrixMutex.Unlock()
}

func closeNetworkConnection() {
	routerAliveConnection.Close()
	routerCommConnection.Close()
}

//Communicating with router functions

func sendToRouter(matrixInTransit map[string]control.ElevatorNode) {
	routerEncoder.Encode(matrixInTransit)

}

func sendToRouterThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		if sendMatrixToRouter {
			elevatorMatrixMutex.Lock()
			tempMatrix := matrixInTransit
			elevatorMatrixMutex.Unlock()
			sendToRouter(tempMatrix)
			sendMatrixToRouter = false
		}
	}
}

func receiveFromRouter() map[string]control.ElevatorNode {
	var receivedData map[string]control.ElevatorNode
	//deadline := time.Now().Add(time.Millisecond * 1)
	//for time.Now().Before(deadline) {}
	routerDecoder.Decode(&receivedData)
	return receivedData
}

func receiveFromRouterThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		if !sendMatrixToRouter {
			tempMatrix := receiveFromRouter()
			elevatorMatrixMutex.Lock()
			matrixInTransit = tempMatrix
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
	for {
		time.Sleep(time.Millisecond * 100)
		text := "Still alive"
		_, err := fmt.Fprintf(routerAliveConnection, text)
		if err != nil {
			return false
		}
		return true
	}
}

func checkRouterStillAlive() bool {
	buf := make([]byte, 1024)
	_, err := routerAliveConnection.Read(buf)
	if err != nil {
		return false
	}
	return true
}

func routerStillAliveThread(initialAddressChannel chan string, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 100)
		if !checkRouterStillAlive() || !tellRouterStillAlive() {
			routerIPAddress = receiveRouterIPFromBackup()
			networkModuleInit(initialAddressChannel, sendToElevatorChannel, receiveFromElevatorChannel)
			time.Sleep(time.Millisecond * 500)
		}
	}
}

//Getting new router info from backup in case router crashes
func receiveRouterIPFromBackup() string {
	/*
		laddr, _ := net.ResolveUDPAddr("udp", net.JoinHostPort(":", backupPort))
		rcv, _ := net.ListenUDP("udp", laddr)
		buff := make([]byte, 1600)
		var str string
		for {
			length, _, _ := rcv.ReadFromUDP(buff)
			str = string(buff[:length])
			if len(str) != 15 { //Assume here 15 is length of IPAddress f.ex. 129.241.187.152
				continue
			} else {
				break
			}
		}
		return string(str)
	*/
	return "129.241.187.153"
}

//Communication with elevator functions
func communicateWithElevatorThread(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	go receiveFromElevatorThread(receiveChannel)
	go sendToElevatorThread(sendChannel)
}

func receiveFromElevatorThread(receiveChannel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 100)
		if !sendMatrixToElevator {
			tempMatrix := <-receiveChannel
			elevatorMatrixMutex.Lock()
			matrixInTransit = tempMatrix
			elevatorMatrixMutex.Unlock()
			sendMatrixToRouter = true
		}
	}
}

func sendToElevatorThread(sendChannel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 100)
		if sendMatrixToElevator {
			elevatorMatrixMutex.Lock()
			tempMatrix := matrixInTransit
			elevatorMatrixMutex.Unlock()
			sendChannel <- tempMatrix
			sendMatrixToElevator = false
		}
	}
}

func Run(initializeAddressChannel chan string, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(3)
	networkModuleInit(initializeAddressChannel, sendToElevatorChannel, receiveFromElevatorChannel)

	go communicateWithElevatorThread(sendToElevatorChannel, receiveFromElevatorChannel)
	go communicateWithRouterThread()
	//go tellRouterStillAliveThread()
	go routerStillAliveThread(initializeAddressChannel, sendToElevatorChannel, receiveFromElevatorChannel)

	wg.Wait()
	closeNetworkConnection()
}
