package control

import (
	"driver"
	"encoding/gob"
	"os"
	"sync"
	"time"
	"user"
)

var elevatorMatrix map[string]ElevatorNode
var elevatorMatrixMutex = &sync.Mutex{}

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

func sendUpdatedMatrix() { //Sends updated map of elevators to network
	openSendChanElevator = true
	if !elevatorIsOffline {
		openSendChanNetwork = true
	}
}

func completePreCrashOrders(orders *ElevatorNode, sendChannel chan map[string]ElevatorNode, receiveChannel chan map[string]ElevatorNode) {
	var ordersMatrix = make(map[string]ElevatorNode)
	LocalAddress = "0"
	for {
		something := *orders
		if ordersEmpty(something) {
			break
		}
		ordersMatrix[LocalAddress] = something
		sendChannel <- ordersMatrix
		ordersMatrix = <-receiveChannel
		Load(driver.BackupOrderFilePath, orders)
	}
}

func controlInit(initializeAddressChannel chan string, blockUserChannel chan bool, blockNetworkChannel chan bool, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode, sendElevatorChannel chan map[string]ElevatorNode, receiveElevatorChannel chan map[string]ElevatorNode) {
	driver.Elev_init() //Initialize hardware
	var tempMatrix = make(map[string]ElevatorNode)
	elevatorMatrix = make(map[string]ElevatorNode)

	var preInitialOrders = new(ElevatorNode)
	err := Load(driver.BackupOrderFilePath, preInitialOrders)
	elevatorHasPreviouslyCrashed := Check(err)
	blockUserChannel <- elevatorHasPreviouslyCrashed
	blockNetworkChannel <- elevatorHasPreviouslyCrashed
	if elevatorHasPreviouslyCrashed {
		completePreCrashOrders(preInitialOrders, sendElevatorChannel, receiveElevatorChannel)
		blockUserChannel <- false
		blockNetworkChannel <- false
	}

	LocalAddress = receiveAddressFromNetwork(initializeAddressChannel)
	LocalElevator := getElevatorState()
	elevatorMatrix[LocalAddress] = LocalElevator
	if LocalAddress == "0" {
		elevatorIsOffline = true
	} else {
		elevatorIsOffline = false
		copyMapByValue(elevatorMatrix, tempMatrix)
		sendNetworkChannel <- tempMatrix
		tempMatrix = <-receiveNetworkChannel
		copyMapByValue(tempMatrix, elevatorMatrix)
	}
}

func distributeOrder(localElevAddress string, newOrder user.ElevatorOrder, elevatorMatrix map[string]ElevatorNode) {
	/*
		for elev, node := range elevatorMatrix {
			fmt.Println(elev)
			fmt.Println(node.CurrentFloor)
		}
	*/
	var tempMatrix = make(map[string]ElevatorNode)
	var bestElevAddress string = localElevAddress //Variable to store best elevator for new order. By default assume initially this is the local elevator
	if newOrder.OrderType == driver.BUTTON_COMMAND {
		//fmt.Println("The order was internal")
		goto ReturnElevator
	} else if newOrder.OrderType == driver.BUTTON_CALL_UP {
		//fmt.Println("The order was up button")
		//Special case: check if any elevators on ordered floor are going upwards
		for address, elevator := range elevatorMatrix {
			//fmt.Println("Checking if any elevator is on same floor as order")
			if elevator.CurrentFloor == newOrder.Floor && elevator.CurrentDirection == driver.DIRN_UP {
				bestElevAddress = address
				goto ReturnElevator
			}
		}
		//fmt.Println("Didn't find any elevators on same floor")
		for i := newOrder.Floor; i >= 0; i-- {
			//fmt.Println("Checking if there are any empty elevators on floor under the order")
			for address, elevator := range elevatorMatrix {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
			//fmt.Println("Checking if there are any elevators going up on floor under the order")
			for address, elevator := range elevatorMatrix {
				if elevator.CurrentFloor == i && elevator.CurrentDirection == driver.DIRN_UP {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
		for i := newOrder.Floor; i <= driver.N_FLOORS; i++ {
			for address, elevator := range elevatorMatrix {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
	} else if newOrder.OrderType == driver.BUTTON_CALL_DOWN {
		//fmt.Println("The order was down button")
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
		for i := newOrder.Floor; i >= 0; i-- {
			for address, elevator := range elevatorMatrix {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
	}
ReturnElevator:
	copyMapByValue(elevatorMatrix, tempMatrix)
	tempElevNode := tempMatrix[bestElevAddress]
	tempElevNode.CurrentOrders[newOrder.OrderType][newOrder.Floor] = true
	elevatorMatrixMutex.Lock()
	tempMatrix[bestElevAddress] = tempElevNode
	copyMapByValue(tempMatrix, elevatorMatrix)
	elevatorMatrixMutex.Unlock()
}



//The application was reliant on being able to copy maps by value, and as ElevatorNode is defined
//in the control module, copyMapByValue() is also defined in the control module
func CopyMapByValue(originalMap map[string]ElevatorNode, newMap map[string]ElevatorNode) {
	for k, _ := range newMap {
		delete(newMap, k)
	}
	for k, v := range originalMap {
		newMap[k] = v
	}
}


func Run(initializeAddressChannel chan string, blockUserChannel chan bool, blockElevatorChannel chan bool, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode, sendElevatorChannel chan map[string]ElevatorNode, receiveElevatorChannel chan map[string]ElevatorNode, receiveUserChannel chan user.ElevatorOrder) {
	wg := new(sync.WaitGroup)
	wg.Add(4)

	controlInit(initializeAddressChannel, blockUserChannel, blockElevatorChannel, sendNetworkChannel, receiveNetworkChannel, sendElevatorChannel, receiveElevatorChannel)

	go networkThread(sendNetworkChannel, receiveNetworkChannel)
	go userThread(receiveUserChannel)
	go elevatorThread(sendElevatorChannel, receiveElevatorChannel)
	go checkConnectedThread(initializeAddressChannel, sendNetworkChannel, receiveNetworkChannel)
	
	wg.Wait()
}
