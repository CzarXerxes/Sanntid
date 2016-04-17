package control

import (
	"driver"
	"sync"
	"user"
)

var elevatorOrderMap map[string]ElevatorNode
var elevatorOrderMapMutex = &sync.Mutex{}

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

func sendUpdatedOrderMap() {
	openSendChanElevator = true
	if !elevatorIsOffline {
		openSendChanNetwork = true
	}
}

func completePreCrashOrders(orders *ElevatorNode, sendChannel chan map[string]ElevatorNode, receiveChannel chan map[string]ElevatorNode) {
	var previousOrderMap = make(map[string]ElevatorNode)
	LocalAddress = "0"
	for {
		something := *orders
		if ordersEmpty(something) {
			break
		}
		previousOrderMap[LocalAddress] = something
		sendChannel <- previousOrderMap
		previousOrderMap = <-receiveChannel
		driver.Load(driver.BackupOrderFilePath, orders)
	}
}

func controlInit(initializeAddressChannel chan string, blockUserChannel chan bool, blockNetworkChannel chan bool, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode, sendElevatorChannel chan map[string]ElevatorNode, receiveElevatorChannel chan map[string]ElevatorNode) {
	driver.Elev_init() 
	var tempOrderMap = make(map[string]ElevatorNode)
	elevatorOrderMap = make(map[string]ElevatorNode)

	var preInitialOrders = new(ElevatorNode)
	err := driver.Load(driver.BackupOrderFilePath, preInitialOrders)
	elevatorHasPreviouslyCrashed := driver.Check(err)
	blockUserChannel <- elevatorHasPreviouslyCrashed
	blockNetworkChannel <- elevatorHasPreviouslyCrashed
	if elevatorHasPreviouslyCrashed {
		completePreCrashOrders(preInitialOrders, sendElevatorChannel, receiveElevatorChannel)
		blockUserChannel <- false
		blockNetworkChannel <- false
	}

	LocalAddress = receiveAddressFromNetwork(initializeAddressChannel)
	LocalElevator := getElevatorState()
	elevatorOrderMap[LocalAddress] = LocalElevator
	if LocalAddress == "0" {
		elevatorIsOffline = true
	} else {
		elevatorIsOffline = false
		CopyMapByValue(elevatorOrderMap, tempOrderMap)
		sendNetworkChannel <- tempOrderMap
		tempOrderMap = <-receiveNetworkChannel
		CopyMapByValue(tempOrderMap, elevatorOrderMap)
	}
}

func distributeOrder(localElevAddress string, newOrder user.ElevatorOrder, elevatorOrderMap map[string]ElevatorNode) {
	var tempOrderMap = make(map[string]ElevatorNode)
	var bestElevAddress string = localElevAddress
	if newOrder.OrderType == driver.BUTTON_COMMAND {
		goto ReturnElevator
	} else if newOrder.OrderType == driver.BUTTON_CALL_UP {
		for address, elevator := range elevatorOrderMap {
			if elevator.CurrentFloor == newOrder.Floor && elevator.CurrentDirection == driver.DIRN_UP {
				bestElevAddress = address
				goto ReturnElevator
			}
		}
		for i := newOrder.Floor; i >= 0; i-- {
			for address, elevator := range elevatorOrderMap {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
			for address, elevator := range elevatorOrderMap {
				if elevator.CurrentFloor == i && elevator.CurrentDirection == driver.DIRN_UP {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
		for i := newOrder.Floor; i <= driver.N_FLOORS; i++ {
			for address, elevator := range elevatorOrderMap {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
	} else if newOrder.OrderType == driver.BUTTON_CALL_DOWN {
		for address, elevator := range elevatorOrderMap {
			if elevator.CurrentFloor == newOrder.Floor && elevator.CurrentDirection == driver.DIRN_DOWN {
				bestElevAddress = address
				goto ReturnElevator
			}
		}
		for i := newOrder.Floor; i <= driver.N_FLOORS; i++ {
			for address, elevator := range elevatorOrderMap {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
			for address, elevator := range elevatorOrderMap {
				if elevator.CurrentFloor == i && elevator.CurrentDirection == driver.DIRN_DOWN {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
		for i := newOrder.Floor; i >= 0; i-- {
			for address, elevator := range elevatorOrderMap {
				if elevator.CurrentFloor == i && ordersEmpty(elevator) {
					bestElevAddress = address
					goto ReturnElevator
				}
			}
		}
	}
ReturnElevator:
	CopyMapByValue(elevatorOrderMap, tempOrderMap)
	tempElevNode := tempOrderMap[bestElevAddress]
	tempElevNode.CurrentOrders[newOrder.OrderType][newOrder.Floor] = true
	elevatorOrderMapMutex.Lock()
	tempOrderMap[bestElevAddress] = tempElevNode
	CopyMapByValue(tempOrderMap, elevatorOrderMap)
	elevatorOrderMapMutex.Unlock()
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
