package control

import (
	"time"
)

var LocalAddress string
var openSendChanNetwork bool = false
var elevatorIsOffline bool = false


func receiveAddressFromNetwork(initializeAddressChannel chan string) string {
	address := <-initializeAddressChannel
	return address
}

func setupElevatorInOnlineMode(tempAddress string, initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	var tempOrderMap = make(map[string]ElevatorNode)
	elevatorIsOffline = false
	openSendChanElevator = false
	CopyMapByValue(elevatorOrderMap, tempOrderMap)
	tempNode := tempOrderMap[LocalAddress]
	CopyMapByValue(tempOrderMap, elevatorOrderMap)
	tempOrderMap = make(map[string]ElevatorNode)
	tempOrderMap[tempAddress] = tempNode
	elevatorOrderMapMutex.Lock()
	CopyMapByValue(tempOrderMap, elevatorOrderMap)
	openSendChanElevator = true
	elevatorOrderMapMutex.Unlock()
	LocalAddress = tempAddress
	CopyMapByValue(elevatorOrderMap, tempOrderMap)
	sendNetworkChannel <- tempOrderMap
	time.Sleep(time.Millisecond * 500)
	tempOrderMap = <-receiveNetworkChannel
	CopyMapByValue(tempOrderMap, elevatorOrderMap)
}

func setupElevatorInOfflineMode(tempAddress string) {
	var tempOrderMap = make(map[string]ElevatorNode)
	elevatorIsOffline = true
	openSendChanElevator = true
	openSendChanNetwork = false
	CopyMapByValue(elevatorOrderMap, tempOrderMap)
	tempNode := tempOrderMap[LocalAddress]
	CopyMapByValue(tempOrderMap, elevatorOrderMap)
	tempOrderMap = make(map[string]ElevatorNode)
	tempOrderMap[tempAddress] = tempNode
	elevatorOrderMapMutex.Lock()
	CopyMapByValue(tempOrderMap, elevatorOrderMap)
	elevatorOrderMapMutex.Unlock()
	LocalAddress = tempAddress
}

func checkConnectedThread(initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	var tempAddress string
	for {
		time.Sleep(time.Millisecond * 10)
		if elevatorIsOffline {
			if len(tempAddress) > 5 {
				setupElevatorInOnlineMode(tempAddress, initializeAddressChannel, sendNetworkChannel, receiveNetworkChannel)
			}
			tempAddress = receiveAddressFromNetwork(initializeAddressChannel)
		} else {
			tempAddress = receiveAddressFromNetwork(initializeAddressChannel)
			if tempAddress == "0" {
				setupElevatorInOfflineMode(tempAddress)
			}
		}
	}
}

func networkThread(sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	go receiveNewOrderMapNetwork(receiveNetworkChannel)
	go sendNewOrderMapNetwork(sendNetworkChannel)
}

func receiveNewOrderMapNetwork(receiveNetworkChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		if !elevatorIsOffline {
			tempOrderMap := <-receiveNetworkChannel
			elevatorOrderMapMutex.Lock()
			if tempOrderMap != nil {
				CopyMapByValue(tempOrderMap, elevatorOrderMap)
			}
			elevatorOrderMapMutex.Unlock()
			openSendChanElevator = true
		}
	}
}

func sendNewOrderMapNetwork(sendNetworkChannel chan map[string]ElevatorNode) {
	var tempOrderMap = make(map[string]ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanNetwork && !elevatorIsOffline {
			elevatorOrderMapMutex.Lock()
			CopyMapByValue(elevatorOrderMap, tempOrderMap)
			elevatorOrderMapMutex.Unlock()
			sendNetworkChannel <- tempOrderMap
			openSendChanNetwork = false

		}
	}
}
