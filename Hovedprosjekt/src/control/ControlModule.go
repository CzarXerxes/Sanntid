package control

import (
	"driver"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"
	"user"
)

var backupOrderFilePath = "/home/student/Desktop/Heis/backupOrders.gob"

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
	newOrder := <-receiveChannel
	return newOrder
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
		Load(backupOrderFilePath, orders)
	}
}

//Functions relating to internal behaviour
func controlInit(initializeAddressChannel chan string, blockUserChannel chan bool, blockNetworkChannel chan bool, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode, sendElevatorChannel chan map[string]ElevatorNode, receiveElevatorChannel chan map[string]ElevatorNode) {
	driver.Elev_init() //Initialize hardware
	var tempMatrix = make(map[string]ElevatorNode)
	elevatorMatrix = make(map[string]ElevatorNode)

	var preInitialOrders = new(ElevatorNode)
	err := Load(backupOrderFilePath, preInitialOrders)
	fmt.Println(err)
	fmt.Println(preInitialOrders)
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

func setupOnline(tempAddress string, initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	var tempMatrix = make(map[string]ElevatorNode)
	elevatorIsOffline = false
	openSendChanElevator = false
	copyMapByValue(elevatorMatrix, tempMatrix)
	tempNode := tempMatrix[LocalAddress]
	copyMapByValue(tempMatrix, elevatorMatrix)
	tempMatrix = make(map[string]ElevatorNode)
	tempMatrix[tempAddress] = tempNode
	elevatorMatrixMutex.Lock()
	copyMapByValue(tempMatrix, elevatorMatrix)
	openSendChanElevator = true
	elevatorMatrixMutex.Unlock()
	LocalAddress = tempAddress
	copyMapByValue(elevatorMatrix, tempMatrix)
	sendNetworkChannel <- tempMatrix
	time.Sleep(time.Millisecond * 500)
	tempMatrix = <-receiveNetworkChannel
	copyMapByValue(tempMatrix, elevatorMatrix)
}

func setupOffline(tempAddress string) {
	var tempMatrix = make(map[string]ElevatorNode)
	//fmt.Println("Before initialized offline mode")
	//fmt.Println(elevatorMatrix)
	elevatorIsOffline = true
	openSendChanElevator = true
	openSendChanNetwork = false
	copyMapByValue(elevatorMatrix, tempMatrix)
	tempNode := tempMatrix[LocalAddress]
	copyMapByValue(tempMatrix, elevatorMatrix)
	tempMatrix = make(map[string]ElevatorNode)
	tempMatrix[tempAddress] = tempNode
	elevatorMatrixMutex.Lock()
	copyMapByValue(tempMatrix, elevatorMatrix)
	elevatorMatrixMutex.Unlock()
	LocalAddress = tempAddress
}

func checkConnectedThread(initializeAddressChannel chan string, sendNetworkChannel chan map[string]ElevatorNode, receiveNetworkChannel chan map[string]ElevatorNode) {
	var tempAddress string
	for {
		time.Sleep(time.Millisecond * 10)
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
					fmt.Println("Found an elevator under floor that was empty", address)
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
	}
	for address, elevator := range elevatorMatrix {
		if ordersEmpty(elevator) {
			bestElevAddress = address
			goto ReturnElevator
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
	//fmt.Println("The new order matrix")
	//fmt.Println(elevatorMatrix)
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
				copyMapByValue(tempMatrix, elevatorMatrix)
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
			copyMapByValue(elevatorMatrix, tempMatrix)
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
		time.Sleep(time.Millisecond * 10)
		newOrder := receiveOrder(receiveChannel)
		fmt.Println("Received this order because someone pushed a button")
		fmt.Println(newOrder)
		distributeOrder(LocalAddress, newOrder, elevatorMatrix)
		sendUpdatedMatrix()
	}

}

func elevatorThread(sendChannel chan map[string]ElevatorNode, receiveChannel chan map[string]ElevatorNode) {
	go receiveNewMatrixElevator(receiveChannel)
	go sendNewMatrixElevator(sendChannel)
}

func receiveNewMatrixElevator(receiveChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		tempMatrix := <-receiveChannel
		elevatorMatrixMutex.Lock()
		copyMapByValue(tempMatrix, elevatorMatrix)
		elevatorMatrixMutex.Unlock()
		if !elevatorIsOffline {
			openSendChanNetwork = true
		}
	}
}

func sendNewMatrixElevator(sendChannel chan map[string]ElevatorNode) {
	var tempMatrix = make(map[string]ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanElevator {
			elevatorMatrixMutex.Lock()
			copyMapByValue(elevatorMatrix, tempMatrix)
			elevatorMatrixMutex.Unlock()
			sendChannel <- tempMatrix
			if !elevatorIsOffline {
				openSendChanElevator = false
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////
//Utility functions
//Put these in their own file later
func copyMapByValue(originalMap map[string]ElevatorNode, newMap map[string]ElevatorNode) {
	for k, _ := range newMap {
		delete(newMap, k)
	}
	for k, v := range originalMap {
		newMap[k] = v
	}
}

func Save(path string, object interface{}) error {
	file, err := os.Create(path)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func Load(path string, object interface{}) error {
	file, err := os.Open(path)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}

func Check(e error) bool {
	if e != nil {
		return false
	}
	return true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////77

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
