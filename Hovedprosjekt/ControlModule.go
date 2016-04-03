package main

import (
	"driver"
	"fmt"
	"sync"
)

var localAddress int

const (
	Downward = -1
	Still    = 0
	Upward   = 1
)

type ElevatorOrder struct {
	orderType driver.Elev_button_type_t //Down order = 0 Up order = 1, Internal order = 2
	floor     int                       //1 indexed(Floor 1 = 1, Floor 2 = 2 ...)
}

type ElevatorNode struct {
	floor         int
	direction     driver.Elev_motor_direction_t
	currentOrders [driver.N_BUTTONS][driver.N_FLOORS]bool
}

//Functions relating to communication with elevator module
func getElevatorState() ElevatorNode { //Get current elevator state from Elevator module
	var elevator ElevatorNode
	elevator.floor = 1
	elevator.direction = Still
	var temp [driver.N_BUTTONS][driver.N_FLOORS]bool
	elevator.currentOrders = temp
	return elevator
}

func updateOrders() {

}

//Functions relating to communication with network module
func requestAddress() int { //Request address from network
	return 0
}

func getOtherElevators() map[int]ElevatorNode { //Get map containing other elevators from network
	m := make(map[int]ElevatorNode)
	return m
}

func sendUpdatedList(elevatorList map[int]ElevatorNode) { //Sends updated map of elevators to network
	fmt.Println(elevatorList)
}

//Functions relating to communication with user module
func receiveOrder() ElevatorOrder {

}

//Functions relating to internal behaviour
func controlInit() map[int]ElevatorNode {
	driver.Elev_init() //Initialize hardware

	elevatorList := getOtherElevators()
	localAddress = requestAddress()
	localElevator := getElevatorState()

	elevatorList[localAddress] = localElevator

	sendUpdatedList(elevatorList)
	return elevatorList
}

func ordersEmpty(elevator ElevatorNode) bool {
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			if elevator.currentOrders[i][j] {
				return false
			}
		}
	}
	return true
}

func distributeOrder(localElevAddress int, newOrder ElevatorOrder, elevatorList map[int]ElevatorNode) map[int]ElevatorNode {
	var bestElevAddress int = localElevAddress //Variable to store best elevator for new order. By default assume initially this is the local elevator
	if newOrder.orderType == driver.BUTTON_COMMAND {
		goto ReturnElevator
	} else if newOrder.orderType == driver.BUTTON_CALL_UP {
		//Special case: check if any elevators on ordered floor are going upwards
		for address, elevator := range elevatorList {
			if elevator.floor == newOrder.floor && elevator.direction == driver.DIRN_UP {
				bestElevAddress = address
				goto ReturnElevator
			}
		}
		for i := newOrder.floor; i >= 0; i-- {
			for address, elevator := range elevatorList {
				if elevator.floor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
			for address, elevator := range elevatorList {
				if elevator.floor == i && elevator.direction == driver.DIRN_UP {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
	} else if newOrder.orderType == driver.BUTTON_CALL_DOWN {
		//Special case: check if any elevators on ordered floor are going downwards
		for address, elevator := range elevatorList {
			if elevator.floor == newOrder.floor && elevator.direction == driver.DIRN_DOWN {
				bestElevAddress = address
				goto ReturnElevator
			}
		}
		for i := newOrder.floor; i <= driver.N_FLOORS; i++ {
			for address, elevator := range elevatorList {
				if elevator.floor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
			for address, elevator := range elevatorList {
				if elevator.floor == i && elevator.direction == driver.DIRN_DOWN {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
	}

ReturnElevator:
	tempElevNode := elevatorList[bestElevAddress]
	tempElevNode.currentOrders[newOrder.orderType][newOrder.floor] = true
	elevatorList[bestElevAddress] = tempElevNode
	return elevatorList
}

func networkThread() {
	for {
		fmt.Println("I am the network!")
	}
}

func userThread() {
	for {
		newOrder := receiveOrder()
		elevatorList = distributeOrder(localAddress, newOrder, elevatorList)
		sendUpdatedList(elevatorList)

	}

}

func elevatorThread() {
	for {
		updateOrders()
	}

}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(3)

	elevatorList := controlInit()
	fmt.Println(elevatorList)

	go networkThread()
	go userThread()
	go elevatorThread()
	wg.Wait()
}
