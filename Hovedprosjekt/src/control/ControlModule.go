package control

import (
	"driver"
	//"fmt"
	"sync"
	"time"
	"user"
)

var elevatorMatrix map[string]ElevatorNode
var LocalAddress string
var openSendChanElevator bool = false
var openSendChanNetwork bool = false

var elevatorIsOffline bool = false

var elevatorMatrixMutex = &sync.Mutex{}

//var openSendChanNetwork bool = false

const (
	Downward = -1
	Still    = 0
	Upward   = 1
)

type ElevatorNode struct {
	CurrentFloor     int
	CurrentDirection driver.Elev_motor_direction_t
	CurrentOrders    [driver.N_BUTTONS][driver.N_FLOORS]bool
}

//Functions relating to communication with elevator module
func getElevatorState() ElevatorNode { //Get current elevator state from Elevator module
	var elevator ElevatorNode
	elevator.CurrentFloor = 1
	elevator.CurrentDirection = Still
	var temp [driver.N_BUTTONS][driver.N_FLOORS]bool
	elevator.CurrentOrders = temp
	return elevator
}

/*
func updateOrders() {

}
*/

//Functions relating to communication with network module
func receiveAddressFromNetwork(initializeAddressChannel chan string) string { //Request address from network
	address := <-initializeAddressChannel
	return address
}

func sendUpdatedMatrix() { //Sends updated map of elevators to network
	openSendChanElevator = true
	if !elevatorIsOffline {
		openSendChanNetwork = true
	}
}

//Functions relating to communication with user module
func receiveOrder(receiveChannel chan user.ElevatorOrder) user.ElevatorOrder {
	for {
		newOrder := <-receiveChannel
		//fmt.Println("Control module : Received an order:")
		//fmt.Println(newOrder)
		return newOrder
	}
}

//Functions relating to internal behaviour
func controlInit(initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	driver.Elev_init() //Initialize hardware
	elevatorMatrix = make(map[string]ElevatorNode)
	LocalAddress = receiveAddressFromNetwork(initializeAddressChannel)
	LocalElevator := getElevatorState()
	elevatorMatrix[LocalAddress] = LocalElevator
	if LocalAddress == "0" {
		elevatorIsOffline = true
	} else {
		elevatorIsOffline = false
		sendNetworkChannel <- elevatorMatrix
		elevatorMatrix = <-receiveNetworkChannel
	}

}

func setupOnline(tempAddress string, initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	elevatorIsOffline = false
	openSendChanElevator = false
	tempNode := elevatorMatrix[LocalAddress]
	tempMatrix := make(map[string]ElevatorNode)
	tempMatrix[tempAddress] = tempNode
	elevatorMatrixMutex.Lock()
	elevatorMatrix = tempMatrix
	openSendChanElevator = true
	elevatorMatrixMutex.Unlock()
	LocalAddress = tempAddress
	sendNetworkChannel <- elevatorMatrix
	time.Sleep(time.Millisecond * 500)
	elevatorMatrix = <-receiveNetworkChannel
}

func setupOffline(tempAddress string) {
	//fmt.Println("Before initialized offline mode")
	//fmt.Println(elevatorMatrix)
	elevatorIsOffline = true
	openSendChanElevator = true
	openSendChanNetwork = false
	tempNode := elevatorMatrix[LocalAddress]
	tempMatrix := make(map[string]ElevatorNode)
	tempMatrix[tempAddress] = tempNode
	elevatorMatrixMutex.Lock()
	elevatorMatrix = tempMatrix
	elevatorMatrixMutex.Unlock()
	//fmt.Println("After initialized offline mode")
	//fmt.Println(elevatorMatrix)
	LocalAddress = tempAddress
}

func checkConnectedThread(initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	var tempAddress string
	for {
		time.Sleep(time.Millisecond * 100)
		if elevatorIsOffline {
			if len(tempAddress) > 5 {
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

/*
func checkConnectedThread(initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	var tempAddress string
	for {elevatorMatrix
		time.Sleep(time.Millisecond * 100)
		if elevatorIsOffline {
			if len(tempAddress) > 5 {
				fmt.Println(tempAddress)
				elevatorIsOffline = false
				openSendChanElevator = false
				tempNode := elevatorMatrix[LocalAddress]
				tempMatrix := make(map[string]ElevatorNode)
				tempMatrix[tempAddress] = tempNode
				elevatorMatrix = tempMatrix
				LocalAddress = tempAddress
				sendNetworkChannel <- elevatorMatrix
				time.Sleep(time.Millisecond * 500)
				elevatorMatrix = <-receiveNetworkChannel
				fmt.Println(elevatorMatrix)
			}
			tempAddress = receiveAddressFromNetwork(initializeAddressChannel)
		} else {
			tempAddress = receiveAddressFromNetwork(initializeAddressChannel)
			if tempAddress == "0" {
				elevatorIsOffline = true
				openSendChanElevator = true
				openSendChanNetwork = false
				tempNode := elevatorMatrix[LocalAddress]
				tempMatrix := make(map[string]ElevatorNode)
				tempMatrix[tempAddress] = tempNode
				elevatorMatrix = tempMatrix
				LocalAddress = tempAddress
			}
		}
	}
}
*/

func ordersEmpty(elevator ElevatorNode) bool {
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			if elevator.CurrentOrders[i][j] {
				return false
			}
		}
	}
	return true
}

func distributeOrder(localElevAddress string, newOrder user.ElevatorOrder, elevatorMatrix map[string]ElevatorNode) {
	var bestElevAddress string = localElevAddress //Variable to store best elevator for new order. By default assume initially this is the local elevator
	if newOrder.OrderType == driver.BUTTON_COMMAND {
		goto ReturnElevator
	} else if newOrder.OrderType == driver.BUTTON_CALL_UP {
		//Special case: check if any elevators on ordered floor are going upwards
		for address, elevator := range elevatorMatrix {
			if elevator.CurrentFloor == newOrder.Floor && elevator.CurrentDirection == driver.DIRN_UP {
				bestElevAddress = address
				goto ReturnElevator
			}
		}
		for i := newOrder.Floor; i >= 0; i-- {
			for address, elevator := range elevatorMatrix {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
			for address, elevator := range elevatorMatrix {
				if elevator.CurrentFloor == i && elevator.CurrentDirection == driver.DIRN_UP {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
	} else if newOrder.OrderType == driver.BUTTON_CALL_DOWN {
		//Special case: check if any elevators on ordered floor are going downwards
		for address, elevator := range elevatorMatrix {
			if elevator.CurrentFloor == newOrder.Floor && elevator.CurrentDirection == driver.DIRN_DOWN {
				bestElevAddress = address
				goto ReturnElevator
			}
		}
		for i := newOrder.Floor; i <= driver.N_FLOORS; i++ {
			for address, elevator := range elevatorMatrix {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
			for address, elevator := range elevatorMatrix {
				if elevator.CurrentFloor == i && elevator.CurrentDirection == driver.DIRN_DOWN {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
	}

ReturnElevator:
	tempElevNode := elevatorMatrix[bestElevAddress]
	tempElevNode.CurrentOrders[newOrder.OrderType][newOrder.Floor] = true
	elevatorMatrixMutex.Lock()
	elevatorMatrix[bestElevAddress] = tempElevNode
	elevatorMatrixMutex.Unlock()
}

func networkThread(sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	go receiveNewMatrixNetwork(receiveNetworkChannel)
	go sendNewMatrixNetwork(sendNetworkChannel)
}

func receiveNewMatrixNetwork(receiveNetworkChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 100)
		if !elevatorIsOffline {
			tempMatrix := <-receiveNetworkChannel
			elevatorMatrixMutex.Lock()
			if tempMatrix != nil {
				elevatorMatrix = tempMatrix
			}
			//fmt.Println("Network thread changed elevatorMatrix to this")
			//fmt.Println(elevatorMatrix)
			elevatorMatrixMutex.Unlock()
			openSendChanElevator = true
		}
	}
}

func sendNewMatrixNetwork(sendNetworkChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanNetwork && !elevatorIsOffline {
			elevatorMatrixMutex.Lock()
			tempMatrix := elevatorMatrix
			elevatorMatrixMutex.Unlock()
			//fmt.Println("Control module : Sending following matrix to network module")
			//fmt.Println(elevatorMatrix)
			sendNetworkChannel <- tempMatrix
			openSendChanNetwork = false

		}
	}
}

func userThread(receiveChannel chan user.ElevatorOrder) {
	for {
		newOrder := receiveOrder(receiveChannel)
		//elevatorMatrixMutex.Lock()
		distributeOrder(LocalAddress, newOrder, elevatorMatrix)
		//fmt.Println("receiveNewMatrixElevator() changed elevatorMatrix to this")
		//fmt.Println(elevatorMatrix)
		//elevatorMatrixMutex.Unlock()
		sendUpdatedMatrix()
	}

}

func elevatorThread(sendChannel chan map[string]ElevatorNode, receiveChannel chan map[string]ElevatorNode) {
	go receiveNewMatrixElevator(receiveChannel)
	go sendNewMatrixElevator(sendChannel)
}

func receiveNewMatrixElevator(receiveChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 100)
		tempMatrix := <-receiveChannel
		elevatorMatrixMutex.Lock()
		elevatorMatrix = tempMatrix
		//fmt.Println("elevatorThread() changed elevatorMatrix to this")
		//fmt.Println(elevatorMatrix)
		elevatorMatrixMutex.Unlock()
		if !elevatorIsOffline {
			openSendChanNetwork = true
		}
	}
}

func sendNewMatrixElevator(sendChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 100)
		if openSendChanElevator {
			elevatorMatrixMutex.Lock()
			tempMatrix := elevatorMatrix
			elevatorMatrixMutex.Unlock()
			sendChannel <- tempMatrix
			if !elevatorIsOffline {
				openSendChanElevator = false
			}
		}
	}
}

func Run(initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode, sendElevatorChannel chan map[string]ElevatorNode, receiveElevatorChannel chan map[string]ElevatorNode, receiveUserChannel chan user.ElevatorOrder) {

	wg := new(sync.WaitGroup)
	wg.Add(4)

	controlInit(initializeAddressChannel, sendNetworkChannel, receiveNetworkChannel)

	go networkThread(sendNetworkChannel, receiveNetworkChannel)
	go userThread(receiveUserChannel)
	go elevatorThread(sendElevatorChannel, receiveElevatorChannel)
	go checkConnectedThread(initializeAddressChannel, sendNetworkChannel, receiveNetworkChannel)
	wg.Wait()
}
