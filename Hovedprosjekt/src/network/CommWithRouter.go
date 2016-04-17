package network

import(
	"net"
	"time"
	"encoding/gob"
	"driver"
	"control"
	"reflect"
)

//Changes with workspace
const IP = "129.241.187.153"
const routerPort = "29000"

var routerIPAddress = IP


func getRouterConnection() bool {
	routerAliveConnection = *new(net.Conn)
	routerCommConnection = *new(net.Conn)
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

unc sendToRouter(matrixInTransit map[string]control.ElevatorNode) {
	var tempData = make(map[string]control.ElevatorNode)
	copyMapByValue(matrixInTransit, tempData)
	routerEncoder.Encode(tempData)

}

func sendToRouterThread() {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if sendMatrixToRouter {
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
				fmt.Println(tempMatrix)
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
			if routerAliveConnection != nil {
				closeNetworkConnection()
			}
			routerIPAddress = nextRouterIP()
			networkModuleInit(false, initialAddressChannel, blockNetworkChan, sendToElevatorChannel, receiveFromElevatorChannel)
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func nextRouterIP() string {
	return IP
}