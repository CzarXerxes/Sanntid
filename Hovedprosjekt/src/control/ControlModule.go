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

func getOtherElevators() map[string]ElevatorNode { //Get map containing other elevators from network
	m := make(map[string]ElevatorNode)
	return m
}

func sendUpdatedMatrix() { //Sends updated map of elevators to network
	openSendChanElevator = true
	openSendChanNetwork = true
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
func controlInit(initializeAddressChannel chan string) {
	driver.Elev_init() //Initialize hardware

	elevatorMatrix = getOtherElevators()
	LocalAddress = receiveAddressFromNetwork(initializeAddressChannel)
	LocalElevator := getElevatorState()

	elevatorMatrix[LocalAddress] = LocalElevator

	//sendUpdatedMatrix(elevatorMatrix)
}

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

func distributeOrder(localElevAddress string, newOrder user.ElevatorOrder, elevatorMatrix map[string]ElevatorNode) map[string]ElevatorNode {
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
	elevatorMatrix[bestElevAddress] = tempElevNode
	return elevatorMatrix
}

func networkThread(sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	go receiveNewMatrixNetwork(receiveNetworkChannel)
	go sendNewMatrixNetwork(sendNetworkChannel)
}

func receiveNewMatrixNetwork(receiveNetworkChannel chan map[string]ElevatorNode) {
	for {
		tempMatrix := <-receiveNetworkChannel
		elevatorMatrixMutex.Lock()
		elevatorMatrix = tempMatrix
		elevatorMatrixMutex.Unlock()
		openSendChanElevator = true
	}
}

func sendNewMatrixNetwork(sendNetworkChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanNetwork {
			elevatorMatrixMutex.Lock()
			//fmt.Println("Control module : Sending following matrix to network module")
			//fmt.Println(elevatorMatrix)
			sendNetworkChannel <- elevatorMatrix
			openSendChanNetwork = false
			elevatorMatrixMutex.Unlock()
		}
	}
}

func userThread(receiveChannel chan user.ElevatorOrder) {
	for {
		newOrder := receiveOrder(receiveChannel)
		elevatorMatrixMutex.Lock()
		elevatorMatrix = distributeOrder(LocalAddress, newOrder, elevatorMatrix)
		elevatorMatrixMutex.Unlock()
		sendUpdatedMatrix()
	}

}

func elevatorThread(sendChannel chan map[string]ElevatorNode, receiveChannel chan map[string]ElevatorNode) {
	go receiveNewMatrixElevator(receiveChannel)
	go sendNewMatrixElevator(sendChannel)
}

func receiveNewMatrixElevator(receiveChannel chan map[string]ElevatorNode) {
	for {
		tempMatrix := <-receiveChannel
		elevatorMatrixMutex.Lock()
		elevatorMatrix = tempMatrix
		elevatorMatrixMutex.Unlock()
		openSendChanNetwork = true
	}
}

func sendNewMatrixElevator(sendChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanElevator {
			elevatorMatrixMutex.Lock()
			//fmt.Println("Control module : Sending following matrix to elevator module")
			//fmt.Println(elevatorMatrix)
			sendChannel <- elevatorMatrix
			openSendChanElevator = false
			elevatorMatrixMutex.Unlock()
		}
	}
}

func Run(initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode, sendElevatorChannel chan map[string]ElevatorNode, receiveElevatorChannel chan map[string]ElevatorNode, receiveUserChannel chan user.ElevatorOrder) {

	wg := new(sync.WaitGroup)
	wg.Add(3)

	controlInit(initializeAddressChannel)

	go networkThread(sendNetworkChannel, receiveNetworkChannel)
	go userThread(receiveUserChannel)
	go elevatorThread(sendElevatorChannel, receiveElevatorChannel)
	wg.Wait()
}
