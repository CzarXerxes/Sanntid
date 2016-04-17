package control


import (
	"time"
)

var LocalAddress string
var openSendChanNetwork bool = false
var elevatorIsOffline bool = false


func receiveAddressFromNetwork(initializeAddressChannel chan string) string { //Request address from network
	address := <-initializeAddressChannel
	return address
}

func setupOnline(tempAddress string, initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	var tempMatrix = make(map[string]ElevatorNode)
	elevatorIsOffline = false
	openSendChanElevator = false
	CopyMapByValue(elevatorMatrix, tempMatrix)
	tempNode := tempMatrix[LocalAddress]
	CopyMapByValue(tempMatrix, elevatorMatrix)
	tempMatrix = make(map[string]ElevatorNode)
	tempMatrix[tempAddress] = tempNode
	elevatorMatrixMutex.Lock()
	CopyMapByValue(tempMatrix, elevatorMatrix)
	openSendChanElevator = true
	elevatorMatrixMutex.Unlock()
	LocalAddress = tempAddress
	CopyMapByValue(elevatorMatrix, tempMatrix)
	sendNetworkChannel <- tempMatrix
	time.Sleep(time.Millisecond * 500)
	tempMatrix = <-receiveNetworkChannel
	CopyMapByValue(tempMatrix, elevatorMatrix)
}

func setupOffline(tempAddress string) {
	var tempMatrix = make(map[string]ElevatorNode)
	elevatorIsOffline = true
	openSendChanElevator = true
	openSendChanNetwork = false
	CopyMapByValue(elevatorMatrix, tempMatrix)
	tempNode := tempMatrix[LocalAddress]
	CopyMapByValue(tempMatrix, elevatorMatrix)
	tempMatrix = make(map[string]ElevatorNode)
	tempMatrix[tempAddress] = tempNode
	elevatorMatrixMutex.Lock()
	CopyMapByValue(tempMatrix, elevatorMatrix)
	elevatorMatrixMutex.Unlock()
	LocalAddress = tempAddress
}

func checkConnectedThread(initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	//var prevConnectedAddress string
	var tempAddress string
	for {
		time.Sleep(time.Millisecond * 10)
		if elevatorIsOffline {
			if len(tempAddress) > 5 {
				//prevConnectedAddress = tempAddress
				setupOnline(tempAddress, initializeAddressChannel, sendNetworkChannel, receiveNetworkChannel)
			}
			tempAddress = receiveAddressFromNetwork(initializeAddressChannel)
		} else {
			tempAddress = receiveAddressFromNetwork(initializeAddressChannel)
			if tempAddress == "0" {
				setupOffline(tempAddress)
			}
		}
	}
}

func networkThread(sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	go receiveNewMatrixNetwork(receiveNetworkChannel)
	go sendNewMatrixNetwork(sendNetworkChannel)
}

func receiveNewMatrixNetwork(receiveNetworkChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		if !elevatorIsOffline {
			tempMatrix := <-receiveNetworkChannel
			elevatorMatrixMutex.Lock()
			if tempMatrix != nil {
				CopyMapByValue(tempMatrix, elevatorMatrix)
			}
			//fmt.Println("Network thread changed elevatorMatrix to this")
			//fmt.Println(elevatorMatrix)
			elevatorMatrixMutex.Unlock()
			openSendChanElevator = true
		}
	}
}

func sendNewMatrixNetwork(sendNetworkChannel chan map[string]ElevatorNode) {
	var tempMatrix = make(map[string]ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanNetwork && !elevatorIsOffline {
			elevatorMatrixMutex.Lock()
			CopyMapByValue(elevatorMatrix, tempMatrix)
			elevatorMatrixMutex.Unlock()
			//fmt.Println("Control module : Sending following matrix to network module")
			//fmt.Println(elevatorMatrix)
			sendNetworkChannel <- tempMatrix
			openSendChanNetwork = false

		}
	}
}