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
const routerIPAddress = "129.241.187.152"

//const port = "20021"
const port = "30000"

//General connection variables
var tcpSendConnection net.Conn

//var tcpReceiveConnection net.Conn
var encoderConnection gob.Encoder
var decoderConnection gob.Decoder

var sendMatrixToRouter bool
var sendMatrixToElevator bool
var routerIsDead bool

var matrixInTransit map[int]control.ElevatorNode

func receiveInitialAddressFromRouter() int {
	var address int
	address = 1
	return address
}

func sendInitialAddressToElevator(address int) {
	fmt.Println(tcpSendConnection.LocalAddr().String())
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
	//wg := new(sync.WaitGroup)
	//wg.Add(1)
	go getTCPSendConnection()
	//go getTCPReceiveConnection()
	//wg.Wait()
	time.Sleep(time.Millisecond * 500)

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
	/*
	go sendToRouterThread()
	go receiveFromRouterThread()
	*/
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

//Thread to tell router module that this elevator is still connected to the network
func tellRouterStillAliveThread() {
	for{		
		time.Sleep(time.Millisecond * 100)
		//fmt.Println("Sending im alive")
		text := "Still alive"
		fmt.Fprintf(tcpSendConnection, text)	
	}
}


func checkRouterStillAliveThread(){
	buf := make([]byte, 1024)
	_, err := tcpSendConnection.Read(buf)
	if err != nil {
		routerIsDead = true
	}
	//fmt.Printf("Message received :: %s\n", string(buf[:n]))
	routerIsDead = false
}

func receiveRouterIPFromBackup() string{
	laddr, _ := net.ResolveUDPAddr("udp", net.JoinHostPort(":", port))
	rcv, _ := net.ListenUDP("udp", laddr)
	buff := make([]byte, 1600)
	var str string
	for{
		length, _, _ := rcv.ReadFromUDP(buff)
		str = string(buff[:length])
		if(len(str) != 15){//Assume here 15 is length of IPAddress f.ex. 129.241.187.152
			continue
		}else{
			break
		}
	}
	return string(str)
	//return "129.241.187.152"
}

func routerIsDeadThread(){
	for{
		if routerIsDead{
			routerIPAddress := receiveRouterIPFromBackup()
			fmt.Println("Router died. Making connection to new router")
			tcpSendConnection, _ = net.Dial("tcp", net.JoinHostPort(routerIPAddress, port))
			fmt.Println("Connected to new router")
			routerIsDead = false
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
	go tellRouterStillAliveThread()
	go checkRouterStillAliveThread()
	go routerIsDeadThread()

	wg.Wait()
	closeNetworkConnection()
}
