package network

import(
	"sync"
	"driver"
	"control"
	"encoding/gob"
	"time"
	"net"
)

var routerAliveConnection net.Conn
var routerCommConnection net.Conn

var routerEncoder *gob.Encoder
var routerDecoder *gob.Decoder

var elevatorMatrixMutex = &sync.Mutex{}

var sendMatrixToRouter bool
var sendMatrixToElevator bool

var matrixInTransit map[string]control.ElevatorNode
var matrixMostRecentlySent map[string]control.ElevatorNode

func getIPAddress() string {
	address := routerAliveConnection.LocalAddr().String()
	return address
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
		time.Sleep(time.Millisecond * 100)
		sendInitialAddressToElevator("0", initializeAddressChannel)
	}
	localAddress := getIPAddress()
	sendInitialAddressToElevator(localAddress, initializeAddressChannel)
	tempMatrix = <-receiveFromElevatorChannel
	sendToRouter(tempMatrix)
	time.Sleep(time.Millisecond * 500)
	tempMatrix = receiveFromRouter()
	elevatorMatrixMutex.Lock()
	control.CopyMapByValue(tempMatrix, matrixInTransit)
	sendToElevatorChannel <- tempMatrix
	elevatorMatrixMutex.Unlock()
}

func closeNetworkConnection() {
	routerAliveConnection.Close()
	routerCommConnection.Close()
}

func Run(initializeAddressChannel chan string, blockNetworkChan chan bool, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(4)
	networkModuleInit(true, initializeAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)

	go communicateWithElevatorThread(sendToElevatorChannel, receiveFromElevatorChannel)
	go communicateWithRouterThread()
	go checkRouterStillAliveThread(initializeAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)
	go tellRouterStillAliveThread(initializeAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)

	wg.Wait()
	closeNetworkConnection()
}
