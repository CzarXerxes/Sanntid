package control

import (
	"driver"
	"fmt"
	"sync"
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
func requestAddress() int { //Request address from network
	return 0
}

func getOtherElevators() map[int]ElevatorNode { //Get map containing other elevators from network
	m := make(map[int]ElevatorNode)
	return m
}

func sendUpdatedMatrix(elevatorMatrix map[int]ElevatorNode) { //Sends updated map of elevators to network
	fmt.Println(elevatorMatrix)
}

//Functions relating to communication with user module
func receiveOrder(c chan user.ElevatorOrder) user.ElevatorOrder {
	for {
		newOrder := <-c
		return newOrder
	}
}

//Functions relating to internal behaviour
func controlInit() {
	driver.Elev_init() //Initialize hardware

	elevatorMatrix = getOtherElevators()
	localAddress = requestAddress()
	localElevator := getElevatorState()

	elevatorMatrix[localAddress] = localElevator

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

/*
func userThread(c chan user.ElevatorOrder, elevatorMatrix map[int]ElevatorNode) {
	for {
		newOrder := receiveOrder(c)
		elevatorMatrix := distributeOrder(localAddress, newOrder, elevatorMatrix)
		sendUpdatedMatrix(elevatorMatrix)
	}

}
*/

func elevatorThread(sendChannel chan map[int]control.ElevatorNode, receiveChannel chan map[int]control.ElevatorNode) {
	go receiveNewMatrixElevator()
	go sendNewMatrixElevator()
}

func receiveNewMatrixElevator(receiveChannel chan map[int]control.ElevatorNode) {
	for {
		elevatorMatrix := <-receiveChannel
	}
}

func sendNewMatrixElevator(sendChannel chan map[int]control.ElevatorNode) {
	for {
		if openSendChanElevator {
			sendChannel <- elevatorMatrix
		}
	}
}

func Run(sendChannel chan map[int]control.ElevatorNode, receiveChannel chan map[int]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	controlInit()

	//go networkThread()
	//go userThread(c, elevatorMatrix)
	go elevatorThread(sendChannel, receiveChannel)
	wg.Wait()
}
