package elevator

import (
	"control"
	"driver"
	//"fmt"
	"sync"
	"time"
)

//See elev.go for enum declarations for use with elev functions

var currentDirection int

const (
	Downward = -1
	Still    = 0
	Upward   = 1
)

var currentFloor int
var isMoving bool = false

const (
	UpIndex       = 0
	DownIndex     = 1
	InternalIndex = 2
)

var openSendChan bool = false
var elevatorMatrix map[int]control.ElevatorNode

//Extend orderArray to have seperate columns for stopping upwards and downwards
var orderArray [2][driver.N_FLOORS]bool               //false = Do not stop, true = Stop
var lightArray [driver.N_BUTTONS][driver.N_FLOORS]int //0 = Do not turn on light; 1 = Turn on light

//Initialization function
func elevatorModuleInit() {
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			lightArray[i][j] = 0
		}
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			orderArray[i][j] = false
		}
	}
	driver.Elev_init()
	for getCurrentFloor() == -1 {
		setDirection(driver.DIRN_DOWN)
	}
	setDirection(driver.DIRN_STOP)
	currentFloor = getCurrentFloor()
	driver.Elev_set_floor_indicator(currentFloor)
	currentDirection = Still
}

//Sensor functions

func getCurrentFloor() int {
	return driver.Elev_get_floor_sensor_signal()
}

//Light functions
func setLights(lightArray [driver.N_BUTTONS][driver.N_FLOORS]int) {
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			driver.Elev_set_button_lamp(driver.Elev_button_type_t(i), j, lightArray[i][j])
		}
	}
}

func getLightArray() [driver.N_BUTTONS][driver.N_FLOORS]int { //Implement differently. Currently just test
	var tempArray [driver.N_BUTTONS][driver.N_FLOORS]int
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			tempArray[i][j] = 0
		}
	}
	return tempArray
}

//Elevator logic functions
/*
type Elev_button_type_t int
const(
	BUTTON_CALL_UP = 0
	BUTTON_CALL_DOWN = 1
	BUTTON_COMMAND = 2
)
*/

//Creates orderArray from elevatorMatrix
func createOrderArray() [2][driver.N_FLOORS]bool {
	var tempArray [2][driver.N_FLOORS]bool //tempArray[0][driver.N_FLOORS] corresponds to orders to complete on way UP, tempArray[1][driver.N_FLOORS] corresponds to orders to complete on way DOWN
	var tempNode control.ElevatorNode
	tempNode = elevatorMatrix[control.LocalAddress]
	//Iterate through orders made with UP and DOWN buttons and place them in corresponding spots in tempArray
	for i := 0; i < 2; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			tempArray[i][j] = tempNode.CurrentOrders[i][j]
		}
	}
	//Place orders made with INTERNAL buttons in tempArray
	if isMoving {
		if currentDirection == Upward {
			for i := currentFloor + 1; i < driver.N_FLOORS; i++ {
				tempArray[UpIndex][i] = tempNode.CurrentOrders[2][i] || tempArray[UpIndex][i]
			}
			for i := 0; i <= currentFloor; i++ {
				tempArray[DownIndex][i] = tempNode.CurrentOrders[2][i] || tempArray[DownIndex][i]
			}
		} else if currentDirection == Downward {
			for i := 0; i < currentFloor; i++ {
				tempArray[DownIndex][i] = tempNode.CurrentOrders[2][i] || tempArray[DownIndex][i]
			}
			for i := currentFloor; i < driver.N_FLOORS; i++ {
				tempArray[UpIndex][i] = tempNode.CurrentOrders[2][i] || tempArray[UpIndex][i]
			}
		}
	} else {
		if currentDirection == Upward {
			tempArray[UpIndex][currentFloor] = tempNode.CurrentOrders[2][currentFloor] || tempArray[UpIndex][currentFloor]
		} else if currentDirection == Downward {
			tempArray[DownIndex][currentFloor] = tempNode.CurrentOrders[2][currentFloor] || tempArray[DownIndex][currentFloor]
		} else if currentDirection == Still {
			tempArray[DownIndex][currentFloor] = tempNode.CurrentOrders[2][currentFloor] || tempArray[DownIndex][currentFloor]
			tempArray[UpIndex][currentFloor] = tempNode.CurrentOrders[2][currentFloor] || tempArray[UpIndex][currentFloor]
		}
		for i := 0; i < currentFloor; i++ {
			tempArray[DownIndex][i] = tempNode.CurrentOrders[2][i] || tempArray[DownIndex][i]
		}
		for i := currentFloor + 1; i < driver.N_FLOORS; i++ {
			tempArray[UpIndex][i] = tempNode.CurrentOrders[2][i] || tempArray[UpIndex][i]
		}
	}
	return tempArray
}

//Accessor and mutator functions for orderArray()
func getOrderArray(directionIndex int, floor int) bool { //directionIndex valid values {UpIndex, DownIndex}
	return orderArray[directionIndex][floor]
}

//Writes orderArray to elevatorMatrix
//Do not use this function anywhere except in setOrderArrayToFalse
func setOrderArray(value bool, directionIndex int, floor int) {
	orderArray[directionIndex][floor] = value

	var tempNode control.ElevatorNode
	tempNode = elevatorMatrix[control.LocalAddress]
	tempNode.CurrentOrders[directionIndex][floor] = value
	tempNode.CurrentOrders[InternalIndex][floor] = value
	elevatorMatrix[control.LocalAddress] = tempNode
	openSendChan = true
}

func setOrderArrayToFalse(directionIndex int, floor int) {
	setOrderArray(false, directionIndex, floor)
}

func deleteOrders() {
	if currentDirection == Upward {
		setOrderArrayToFalse(UpIndex, currentFloor)
	} else if currentDirection == Downward {
		setOrderArrayToFalse(DownIndex, currentFloor)
	} else if currentDirection == Still {
		setOrderArrayToFalse(DownIndex, currentFloor)
		setOrderArrayToFalse(UpIndex, currentFloor)
	}
}

func noPendingOrders() bool {
	for i := 0; i < driver.N_FLOORS; i++ {
		if getOrderArray(DownIndex, i) && getOrderArray(UpIndex, i) {
			return false
		}
	}
	return true
}

func calculateCurrentDirection() int { //Finds new currentDirection(Upward,Downward or Still) based on currentDirection and pending orders
	if noPendingOrders() {
		return Still
	}
	switch currentDirection {
	case Still:
		for i := 0; i < driver.N_FLOORS; i++ {
			if getOrderArray(UpIndex, i) || getOrderArray(DownIndex, i) {
				if i == currentFloor {
					return Still
				} else if i < currentFloor {
					return Downward
				} else if i > currentFloor {
					return Upward
				}
			}
		}
	case Upward:
		for i := currentFloor; i < driver.N_FLOORS; i++ {
			if getOrderArray(UpIndex, i) {
				return Upward
			}
		}
		for i := 0; i < currentFloor; i++ {
			if getOrderArray(DownIndex, i) || getOrderArray(UpIndex, i) {
				return Downward
			}
		}
	case Downward:
		for i := 0; i <= currentFloor; i++ {
			if getOrderArray(DownIndex, i) {
				return Downward
			}
		}
		for i := currentFloor + 1; i < driver.N_FLOORS; i++ {
			if getOrderArray(DownIndex, i) || getOrderArray(UpIndex, i) {
				return Upward
			}
		}
	}
	return Still
}

//Elevator movement functions
func setDirection(direction driver.Elev_motor_direction_t) {
	driver.Elev_set_motor_direction(direction)
	if direction != driver.DIRN_STOP {
		isMoving = true
	} else {
		isMoving = false
	}
}

func moveElevator(direction driver.Elev_motor_direction_t) {
	for getCurrentFloor() != -1 {
		setDirection(direction)
	}
	for getCurrentFloor() == -1 {
	}
	currentFloor = getCurrentFloor()
}

func stopElevator() { //Stop elevator, open doors for 5 sec, repeats if more orders have come on same floor in that time
	for getOrderArray(currentDirection, currentFloor) {
		deleteOrders()
		setDirection(driver.DIRN_STOP)
		driver.Elev_set_door_open_lamp(1)
		time.Sleep(time.Second * 5)
	}
	driver.Elev_set_door_open_lamp(0)
}

func floorIsReached() {
	driver.Elev_set_floor_indicator(currentFloor)
	stopElevator()
	currentDirection = calculateCurrentDirection()
}

//Main threads
func lightThread() {
	for {
		setLights(getLightArray())
	}
}

func elevatorMovementThread() {
	for {
		switch currentDirection {
		case Still:
			if getOrderArray(Upward, currentFloor) || getOrderArray(Downward, currentFloor) {
				floorIsReached()
			}
			currentDirection = calculateCurrentDirection()

		case Downward:
			if getOrderArray(Downward, currentFloor) {
				floorIsReached()
			}
			moveElevator(driver.DIRN_DOWN)
		case Upward:
			if getOrderArray(Upward, currentFloor) {
				floorIsReached()
			}
			moveElevator(driver.DIRN_UP)
		default:
			setDirection(driver.DIRN_STOP)
		}
	}
}

func communicationThread(sendChannel chan map[int]control.ElevatorNode, receiveChannel chan map[int]control.ElevatorNode) {
	go receiveNewMatrix(receiveChannel)
	go sendNewMatrix(sendChannel)
}

func receiveNewMatrix(receiveChannel chan map[int]control.ElevatorNode) {
	for {
		elevatorMatrix = <-receiveChannel
		orderArray = createOrderArray()
	}
}

func sendNewMatrix(sendChannel chan map[int]control.ElevatorNode) {
	for {
		if openSendChan {
			sendChannel <- elevatorMatrix
			openSendChan = false
		}
	}
}

func Run(sendChannel chan map[int]control.ElevatorNode, receiveChannel chan map[int]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	elevatorModuleInit()

	go lightThread()
	go elevatorMovementThread()
	go communicationThread(sendChannel, receiveChannel)
	wg.Wait()
}
