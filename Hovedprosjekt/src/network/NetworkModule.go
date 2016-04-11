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
const routerIPAddress = "129.241.187.153"

//const port = "20021"
const routerPort = "29000"
const backupPort = "28000"

//General connection variables
var tcpSendConnection net.Conn

//var tcpReceiveConnection net.Conn

var elevatorMatrixMutex = &sync.Mutex{}
var sendMatrixToRouter bool
var sendMatrixToElevator bool
var routerIsDead bool

var matrixInTransit map[string]control.ElevatorNode

func receiveInitialAddressFromRouter() string {
	address := tcpSendConnection.LocalAddr().String()
	return address
}

func sendInitialAddressToElevator(address string, initializeAddressChannel chan string) {
	initializeAddressChannel <- address
}

func getTCPSendConnection() {
	//fmt.Println("Network module : Making connection to router")
	tcpSendConnection, _ = net.Dial("tcp", net.JoinHostPort(routerIPAddress, routerPort))
	//fmt.Println("Network module : Made connection")
}

func networkModuleInit(initializeAddressChannel chan string) {
	go getTCPSendConnection()
	time.Sleep(time.Millisecond * 500)

	//Return IP address
	localAddress := receiveInitialAddressFromRouter()
	//Send assigned address to elevator
	sendInitialAddressToElevator(localAddress, initializeAddressChannel)
}

func closeNetworkConnection() {
	tcpSendConnection.Close()
}

//Communicating with router functions

func sendToRouter(matrixInTransit map[string]control.ElevatorNode) {
	//fmt.Println("Network module: Sending matrix to router")
	//fmt.Println(matrixInTransit)
	for i := 0; i < 10; i++ {
		//time.Sleep(time.Millisecond * 10)
		enc := gob.NewEncoder(tcpSendConnection)
		//fmt.Println(matrixInTransit)
		enc.Encode(matrixInTransit)
	}

}

func sendToRouterThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		if sendMatrixToRouter {
			//fmt.Println(matrixInTransit)
			sendToRouter(matrixInTransit)
			sendMatrixToRouter = false
			elevatorMatrixMutex.Unlock()
		}
	}
}

func receiveFromRouter() map[string]control.ElevatorNode {
	var receivedData map[string]control.ElevatorNode
	deadline := time.Now().Add(time.Millisecond * 1)
	for time.Now().Before(deadline) {
		dec := gob.NewDecoder(tcpSendConnection)
		dec.Decode(&receivedData)
	}
	return receivedData
}

func receiveFromRouterThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		elevatorMatrixMutex.Lock()
		tempMatrix := receiveFromRouter()
		fmt.Println("Received this from router")
		fmt.Println(tempMatrix)
		matrixInTransit = tempMatrix
		sendMatrixToElevator = true
	}
}

func communicateWithRouterThread() {
	go sendToRouterThread()
	go receiveFromRouterThread()
}

//Thread to tell router module that this elevator is still connected to the network
func tellRouterStillAliveThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		text := "Still alive"
		fmt.Fprintf(tcpSendConnection, text)
	}
}

func routerStillAlive() bool {
	buf := make([]byte, 1024)
	_, err := tcpSendConnection.Read(buf)
	if err != nil {
		return false
	}
	//fmt.Printf("Message received :: %s\n", string(buf[:n]))
	return true
}

func routerStillAliveThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		if !routerStillAlive() {
			//fmt.Println("Network module : Router died. Making connection to new router")
			routerIPAddress := receiveRouterIPFromBackup()
			tcpSendConnection, _ = net.Dial("tcp", net.JoinHostPort(routerIPAddress, routerPort))
			//fmt.Println("Network module : Connected to new router")
		}
	}
}

//Getting new router info from backup in case router crashes
func receiveRouterIPFromBackup() string {
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
	//return "129.241.187.152"
}

//Communication with elevator functions
func communicateWithElevatorThread(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	go receiveFromElevatorThread(receiveChannel)
	go sendToElevatorThread(sendChannel)
}

func receiveFromElevatorThread(receiveChannel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 100)
		tempMatrix := <-receiveChannel
		//fmt.Println("Network module : Received matrix from control module")
		//fmt.Println(tempMatrix)
		elevatorMatrixMutex.Lock()
		matrixInTransit = tempMatrix
		sendMatrixToRouter = true
	}
}

func sendToElevatorThread(sendChannel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 100)
		if sendMatrixToElevator {
			sendChannel <- matrixInTransit
			sendMatrixToElevator = false
			elevatorMatrixMutex.Unlock()
		}
	}
}

func Run(initializeAddressChannel chan string, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(4)
	networkModuleInit(initializeAddressChannel)

	go communicateWithElevatorThread(sendToElevatorChannel, receiveFromElevatorChannel)
	go communicateWithRouterThread()
	go tellRouterStillAliveThread()
	go routerStillAliveThread()

	wg.Wait()
	closeNetworkConnection()
}
