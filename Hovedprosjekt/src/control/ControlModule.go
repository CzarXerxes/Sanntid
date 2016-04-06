package control

import (
	"driver"
	"fmt"
	"sync"
	"time"
	"user"
)

var elevatorMatrix map[int]ElevatorNode
var LocalAddress int
var openSendChanElevator bool = false

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
func receiveAddressFromNetwork(initializeAddressChannel chan int) int { //Request address from network
	address := <-initializeAddressChannel
	return address
}

func getOtherElevators() map[int]ElevatorNode { //Get map containing other elevators from network
	m := make(map[int]ElevatorNode)
	return m
}

func sendUpdatedMatrix() { //Sends updated map of elevators to network
	openSendChanElevator = true
}

//Functions relating to communication with user module
func receiveOrder(receiveChannel chan user.ElevatorOrder) user.ElevatorOrder {
	for {
		newOrder := <-receiveChannel
		fmt.Println("Received an order:")
		fmt.Println(newOrder)
		return newOrder
	}
}

//Functions relating to internal behaviour
func controlInit(initializeAddressChannel chan int) {
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

func distributeOrder(localElevAddress int, newOrder user.ElevatorOrder, elevatorMatrix map[int]ElevatorNode) map[int]ElevatorNode {
	var bestElevAddress int = localElevAddress //Variable to store best elevator for new order. By default assume initially this is the local elevator
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

/*
func networkThread() {
	for {
		//fmt.Println("I am the network!")
	}
}

func receiveNewMatrixNetwork(){

}

func sendNewMatrixNetwork(){

}

*/

func userThread(receiveChannel chan user.ElevatorOrder) {
	for {
		newOrder := receiveOrder(receiveChannel)
		elevatorMatrix = distributeOrder(LocalAddress, newOrder, elevatorMatrix)
		sendUpdatedMatrix()
	}

}

func elevatorThread(sendChannel chan map[int]ElevatorNode, receiveChannel chan map[int]ElevatorNode) {
	go receiveNewMatrixElevator(receiveChannel)
	go sendNewMatrixElevator(sendChannel)
}

func receiveNewMatrixElevator(receiveChannel chan map[int]ElevatorNode) {
	for {
		elevatorMatrix = <-receiveChannel
	}
}

func sendNewMatrixElevator(sendChannel chan map[int]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanElevator {
			fmt.Println("Sending following matrix to elevator module")
			fmt.Println(elevatorMatrix)
			sendChannel <- elevatorMatrix
			openSendChanElevator = false
		}
	}
}

func Run(initializeAddressChannel chan int, sendElevatorChannel chan map[int]ElevatorNode, receiveElevatorChannel chan map[int]ElevatorNode, receiveUserChannel chan user.ElevatorOrder) {

	wg := new(sync.WaitGroup)
	wg.Add(2)

	controlInit(initializeAddressChannel)

	//go networkThread()
	go userThread(receiveUserChannel)
	go elevatorThread(sendElevatorChannel, receiveElevatorChannel)
	wg.Wait()
}
