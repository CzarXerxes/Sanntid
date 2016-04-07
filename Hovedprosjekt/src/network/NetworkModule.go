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

const routerIPAddress = "129.241.187.153"

//const port = "20021"
const port = "30000"

//General connection variables
var tcpSendConnection net.Conn

//var tcpReceiveConnection net.Conn
var encoderConnection gob.Encoder
var decoderConnection gob.Decoder

var sendMatrixToRouter bool
var sendMatrixToElevator bool

var matrixInTransit map[int]control.ElevatorNode

func receiveInitialAddressFromRouter() int {
	var address int
	decoderConnection.Decode(address)
	return address
}

func sendInitialAddressToElevator(address int) {
	fmt.Println(address)
}

/*
func sendInitialAddressToElevator(address int, initializeAddressChannel chan int) {
	initializeAddressChannel <- address
}
*/

func getTCPSendConnection() {
	fmt.Println("Making connection")
	tcpSendConnection, _ = net.Dial("tcp", net.JoinHostPort(routerIPAddress, port))
	fmt.Println("Made connection")
}

/*
func getTCPReceiveConnection() {
	ln, _ := net.Listen("tcp", port)
	tcpReceiveConnection, _ = ln.Accept()
}
*/
func networkModuleInit( /*initializeAddressChannel chan int*/ ) {
	//Create TCP connection to router
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go getTCPSendConnection()
	//go getTCPReceiveConnection()
	wg.Wait()

	encoderConnection = *gob.NewEncoder(tcpSendConnection)
	decoderConnection = *gob.NewDecoder(tcpSendConnection)

	//Receive assigned local address from router

	localAddress := receiveInitialAddressFromRouter()
	//Send assigned address to elevator
	sendInitialAddressToElevator(localAddress /*, initializeAddressChannel*/)
}

func closeNetworkConnection() {
	tcpSendConnection.Close()
	//tcpReceiveConnection.Close()
}

func sendToRouter(matrixInTransit map[int]control.ElevatorNode) {
	encoderConnection.Encode(matrixInTransit)
}

func receiveFromRouter() map[int]control.ElevatorNode {
	receivedData := &map[int]control.ElevatorNode{}
	decoderConnection.Decode(receivedData)
	return *receivedData
}

func sendToRouterThread() {
	for {
		time.Sleep(time.Millisecond * 10)
		if sendMatrixToRouter {
			sendToRouter(matrixInTransit)
			sendMatrixToRouter = false
		}
	}
}

func receiveFromRouterThread() {
	for {
		for sendMatrixToElevator {
		}
		matrixInTransit = receiveFromRouter()
		sendMatrixToElevator = true
	}
}

func communicateWithRouterThread() {
	go sendToRouterThread()
	go receiveFromRouterThread()
}

//Thread to tell router module that this elevator is still connected to the network
func aliveThread() {
	raddr, _ := net.ResolveUDPAddr("udp", net.JoinHostPort(routerIPAddress, port))
	conn, _ := net.DialUDP("udp", nil, raddr)
	defer conn.Close()
	for {
		fmt.Fprintf(conn, string(uint64(control.LocalAddress)))
		time.Sleep(time.Millisecond * 100)
	}
}

func communicateWithElevatorThread() {
	for {
		time.Sleep(time.Millisecond * 10)
		if sendMatrixToElevator {
			fmt.Println(matrixInTransit)
			sendMatrixToElevator = false
		}
	}
}

/*
//Communication with elevator functions
func communicateWithElevatorThread(receiveChannel chan map[int]control.ElevatorNode, sendChannel chan map[int]control.ElevatorNode) {
	go receiveFromElevatorThread(receiveChannel)
	go sendToElevatorThread(sendChannel)
}

func receiveFromElevatorThread(receiveChannel chan map[int]control.ElevatorNode) {
	for {
		tempMatrix := <-receiveChannel
		for sendMatrixToRouter {
		}
		matrixInTransit = tempMatrix
		sendMatrixToRouter = true
	}
}

func sendToElevatorThread(sendChannel chan map[int]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		if sendMatrixToElevator {
			sendChannel <- matrixInTransit
			sendMatrixToElevator = false
		}
	}
}
*/

func Run( /*initializeAddressChannel chan int, receiveFromElevatorChannel chan map[int]control.ElevatorNode, sendToElevatorChannel chan map[int]control.ElevatorNode*/ ) {
	wg := new(sync.WaitGroup)
	wg.Add(3)
	networkModuleInit( /*initializeAddressChannel*/ )

	go communicateWithElevatorThread( /*receiveFromElevatorChannel, sendToElevatorChannel*/ )
	go communicateWithRouterThread()
	go aliveThread()

	wg.Wait()
	closeNetworkConnection()
}
