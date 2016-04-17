package network

import(
	"sync"
	"control"
	"encoding/gob"
	"time"
	"net"
)

var routerAliveConnection net.Conn
var routerCommConnection net.Conn

var routerEncoder *gob.Encoder
var routerDecoder *gob.Decoder

var elevatorOrderMapMutex = &sync.Mutex{}

var sendOrderMapToRouter bool
var sendOrderMapToElevator bool

var orderMapInTransit map[string]control.ElevatorNode
var orderMapMostRecentlySent map[string]control.ElevatorNode

func getIPAddress() string {
	address := routerAliveConnection.LocalAddr().String()
	return address
}

func networkModuleInit(firstTimeCalled bool, initializeAddressChannel chan string, blockNetworkChan chan bool, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	var tempOrderMap = make(map[string]control.ElevatorNode)
	orderMapInTransit = make(map[string]control.ElevatorNode)
	orderMapMostRecentlySent = make(map[string]control.ElevatorNode)
	if firstTimeCalled {
		waitBecauseElevatorsHavePreviouslyCrashed := <-blockNetworkChan
		if waitBecauseElevatorsHavePreviouslyCrashed {
			waitBecauseElevatorsHavePreviouslyCrashed = <-blockNetworkChan
		}
	}

	for !getRouterConnection() {
		time.Sleep(time.Millisecond * 100)
		sendInitialAddressToControlModule("0", initializeAddressChannel)
	}
	localAddress := getIPAddress()
	sendInitialAddressToControlModule(localAddress, initializeAddressChannel)
	tempOrderMap = <-receiveFromElevatorChannel
	sendOrderToRouter(tempOrderMap)
	time.Sleep(time.Millisecond * 500)
	tempOrderMap = receiveOrderFromRouter()
	elevatorOrderMapMutex.Lock()
	control.CopyMapByValue(tempOrderMap, orderMapInTransit)
	sendToElevatorChannel <- tempOrderMap
	elevatorOrderMapMutex.Unlock()
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
