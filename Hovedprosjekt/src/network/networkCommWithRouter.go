package network

import(
	"net"
	"time"
	"encoding/gob"
	"driver"
	"control"
	"reflect"
	"fmt"
)

var routerIPAddress = driver.IP

func getRouterConnection() bool {
	routerAliveConnection = *new(net.Conn)
	routerCommConnection = *new(net.Conn)
	var err error
	routerAliveConnection, err = net.Dial("tcp", net.JoinHostPort(routerIPAddress, driver.Port))
	if err != nil {
		return false
	}
	routerCommConnection, err = net.Dial("tcp", net.JoinHostPort(routerIPAddress, driver.Port))
	if err != nil {
		return false
	}
	routerEncoder = gob.NewEncoder(routerCommConnection)
	routerDecoder = gob.NewDecoder(routerCommConnection)
	return true
}

func sendOrderToRouter(orderMapInTransit map[string]control.ElevatorNode) {
	var tempData = make(map[string]control.ElevatorNode)
	control.CopyMapByValue(orderMapInTransit, tempData)
	routerEncoder.Encode(tempData)

}

func sendOrderToRouterThread() {
	var tempOrderMap = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if sendOrderMapToRouter {
			elevatorOrderMapMutex.Lock()
			control.CopyMapByValue(orderMapInTransit, tempOrderMap)
			control.CopyMapByValue(orderMapInTransit, orderMapMostRecentlySent)
			elevatorOrderMapMutex.Unlock()
			sendOrderToRouter(tempOrderMap)
			sendOrderMapToRouter = false
		}
	}
}

func receiveOrderFromRouter() map[string]control.ElevatorNode {
	var receivedData = make(map[string]control.ElevatorNode)
	var tempData = make(map[string]control.ElevatorNode)
	routerDecoder.Decode(&receivedData)
	control.CopyMapByValue(receivedData, tempData)
	return tempData
}

func receiveOrderFromRouterThread() {
	var tempOrderMap = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if !sendOrderMapToRouter {
			tempOrderMap = receiveOrderFromRouter()
			elevatorOrderMapMutex.Lock()
			if !reflect.DeepEqual(tempOrderMap, orderMapMostRecentlySent) {
				control.CopyMapByValue(tempOrderMap, orderMapInTransit)
			}
			sendOrderMapToElevator = true
			elevatorOrderMapMutex.Unlock()
		}
	}
}

func communicateWithRouterThread() {
	go sendOrderToRouterThread()
	go receiveOrderFromRouterThread()
}

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
			if routerAliveConnection != nil {
				closeNetworkConnection()
			}
			networkModuleInit(false, initialAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func checkRouterStillAliveThread(initialAddressChannel chan string, blockNetworkChan chan bool, sendToElevatorChannel chan map[string]control.ElevatorNode, receiveFromElevatorChannel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 500)
		if !checkRouterStillAlive() {
			if routerAliveConnection != nil {
				closeNetworkConnection()
			}
			networkModuleInit(false, initialAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)
			time.Sleep(time.Millisecond * 500)
		}
	}
}
