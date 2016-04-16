package elevator

import (
	"control"
	"driver"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"
)

//See elev.go for enum declarations for use with elev functions
var backupOrderFilePath = "/home/student/Desktop/Heis/backupOrders.gob"

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

var receivedFirstMatrix bool = false
var openSendChan bool = false
var elevatorMatrix map[string]control.ElevatorNode
var matrixBeingHandled map[string]control.ElevatorNode

var elevatorMatrixMutex = &sync.Mutex{}

//Extend orderArray to have seperate columns for stopping upwards and downwards
var orderArray [2][driver.N_FLOORS]bool               //false = Do not stop, true = Stop
var lightArray [driver.N_BUTTONS][driver.N_FLOORS]int //0 = Do not turn on light; 1 = Turn on light

//Initialization function
func elevatorModuleInit() {
	elevatorMatrix = make(map[string]control.ElevatorNode)
	matrixBeingHandled = make(map[string]control.ElevatorNode)
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

	floor := getCurrentFloor()
	for floor == -1 {
		setDirection(driver.DIRN_DOWN)
		floor = getCurrentFloor()
	}
	setDirection(driver.DIRN_STOP)
	currentFloor = floor
	driver.Elev_set_floor_indicator(currentFloor)
	currentDirection = Still
}

//Sensor functions

func getCurrentFloor() int {
	return driver.Elev_get_floor_sensor_signal()
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
	var tempMatrix = make(map[string]control.ElevatorNode)
	copyMapByValue(elevatorMatrix, tempMatrix)

	tempNode = tempMatrix[control.LocalAddress]
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

func setElevatorMatrixDirection(direction driver.Elev_motor_direction_t) {
	elevatorMatrixMutex.Lock()
	var tempMatrix = make(map[string]control.ElevatorNode)
	copyMapByValue(elevatorMatrix, tempMatrix)
	tempNode := tempMatrix[control.LocalAddress]
	tempNode.CurrentDirection = direction
	tempMatrix[control.LocalAddress] = tempNode
	copyMapByValue(tempMatrix, elevatorMatrix)
	elevatorMatrixMutex.Unlock()
}

func setElevatorMatrixFloor(floor int) {
	elevatorMatrixMutex.Lock()
	var tempMatrix = make(map[string]control.ElevatorNode)
	copyMapByValue(elevatorMatrix, tempMatrix)
	tempNode := tempMatrix[control.LocalAddress]
	tempNode.CurrentFloor = floor
	tempMatrix[control.LocalAddress] = tempNode
	copyMapByValue(tempMatrix, elevatorMatrix)
	elevatorMatrixMutex.Unlock()
}

//Accessor and mutator functions for orderArray()
func getOrderArray(directionIndex int, floor int) bool { //directionIndex valid values {UpIndex, DownIndex}
	return orderArray[directionIndex][floor]
}

func noPendingOrdersDirectionOverFloor(directionIndex int, floor int) bool {
	for i := floor; i < driver.N_FLOORS; i++ {
		if getOrderArray(directionIndex, i) {
			return false
		}
	}
	return true
}

func noPendingOrdersDirectionUnderFloor(directionIndex int, floor int) bool {
	for i := floor; i > 0; i-- {
		if getOrderArray(directionIndex, i) {
			return false
		}
	}
	return true
}

func noPendingOrdersDirection(directionIndex int) bool {
	for i := 0; i < driver.N_FLOORS; i++ {
		if getOrderArray(directionIndex, i) {
			return false
		}
	}
	return true
}

func noPendingOrders() bool {
	return noPendingOrdersDirection(DownIndex) && noPendingOrdersDirection(UpIndex)
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
		time.Sleep(time.Millisecond * 10)
	}
	for getCurrentFloor() == -1 {
		time.Sleep(time.Millisecond * 10)
	}
	tempFloor := currentFloor
	currentFloor = getCurrentFloor()
	if tempFloor != currentFloor {
		setElevatorMatrixFloor(currentFloor)
	}
}

//Main threads
func lightThread() {
	for {
		time.Sleep(time.Millisecond * 10)
		elevatorMatrixMutex.Lock()
		setLights(getLightArray())
		elevatorMatrixMutex.Unlock()
	}
}

//Light functions
func setLights(lightArray [driver.N_BUTTONS][driver.N_FLOORS]int) {
	driver.Elev_set_floor_indicator(currentFloor)
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			driver.Elev_set_button_lamp(driver.Elev_button_type_t(i), j, lightArray[i][j])
		}
	}
}

func getLightArray() [driver.N_BUTTONS][driver.N_FLOORS]int { //Implement differently. Currently just test
	var tempMatrix = make(map[string]control.ElevatorNode)
	var tempArray [driver.N_BUTTONS][driver.N_FLOORS]int
	copyMapByValue(elevatorMatrix, tempMatrix)
	for j := 0; j < driver.N_FLOORS; j++ {
		localOrders := tempMatrix[control.LocalAddress]
		tempArray[2][j] = BoolToInt(localOrders.CurrentOrders[2][j])
		for i := 0; i < driver.N_BUTTONS-1; i++ {
			for _, matrix := range tempMatrix {
				tempArray[i][j] = BoolToInt(matrix.CurrentOrders[i][j] || IntToBool(tempArray[i][j]))
			}
		}
	}
	return tempArray
}

func BoolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func IntToBool(i int) bool {
	if i == 1 {
		return true
	} else {
		return false
	}
}

func setOrderArray(value bool, directionIndex int, floor int) {
	elevatorMatrixMutex.Lock()
	var tempMatrix = make(map[string]control.ElevatorNode)
	copyMapByValue(elevatorMatrix, tempMatrix)

	orderArray[directionIndex][floor] = value

	var tempNode control.ElevatorNode
	tempNode = tempMatrix[control.LocalAddress]
	tempNode.CurrentOrders[directionIndex][floor] = value
	tempNode.CurrentOrders[InternalIndex][floor] = value
	tempMatrix[control.LocalAddress] = tempNode
	copyMapByValue(tempMatrix, elevatorMatrix)
	openSendChan = true
	elevatorMatrixMutex.Unlock()
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

func stopElevator() {
	setDirection(driver.DIRN_STOP)
	driver.Elev_set_door_open_lamp(1)
	time.Sleep(time.Second * 3)
	driver.Elev_set_door_open_lamp(0)
}

func floorIsReached() {
	//driver.Elev_set_floor_indicator(currentFloor)

	stopElevator()
	deleteOrders()
	currentDirection = calculateCurrentDirection()
	deleteOrders()
	//fmt.Println("Deleting orders before looping")
	setElevatorMatrixDirection(driver.Elev_motor_direction_t(currentDirection))
}

func elevatorMovementThread() {
	for {
		//setElevatorMatrixFloor(currentFloor)
		time.Sleep(time.Millisecond * 10)
		if receivedFirstMatrix {
			switch currentDirection {
			case Still:
				fmt.Println("Still state")
				setElevatorMatrixDirection(driver.Elev_motor_direction_t(currentDirection))
				if getOrderArray(UpIndex, currentFloor) || getOrderArray(DownIndex, currentFloor) {
					floorIsReached()
				}
				//elevatorMatrixMutex.Lock()
				currentDirection = calculateCurrentDirection()
				deleteOrders()
				//elevatorMatrixMutex.Unlock()
				//time.Sleep(time.Second)
			case Downward:
				fmt.Println("Down state")
				setElevatorMatrixDirection(driver.Elev_motor_direction_t(currentDirection))
				if noPendingOrdersDirection(DownIndex) {
					if getOrderArray(UpIndex, currentFloor) || getOrderArray(DownIndex, currentFloor) {
						floorIsReached()
					}
				} else {
					if getOrderArray(DownIndex, currentFloor) {
						floorIsReached()
					}
					if noPendingOrdersDirectionUnderFloor(DownIndex, currentFloor) {
						if getOrderArray(UpIndex, currentFloor) {
							floorIsReached()
						}
					}
				}
				if currentDirection == Downward {
					moveElevator(driver.DIRN_DOWN)
				}
			case Upward:
				fmt.Println("Up state")
				setElevatorMatrixDirection(driver.Elev_motor_direction_t(currentDirection))
				if noPendingOrdersDirection(UpIndex) {
					if getOrderArray(UpIndex, currentFloor) || getOrderArray(DownIndex, currentFloor) {
						floorIsReached()
					}
				} else {
					if getOrderArray(UpIndex, currentFloor) {
						floorIsReached()
					}
					if noPendingOrdersDirectionOverFloor(UpIndex, currentFloor) {
						if getOrderArray(DownIndex, currentFloor) {
							floorIsReached()
						}
					}
				}
				if currentDirection == Upward {
					moveElevator(driver.DIRN_UP)
				}
			default:
				setDirection(driver.DIRN_STOP)
			}
		}
	}
}

func communicationThread(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	go receiveNewMatrix(receiveChannel)
	go sendNewMatrix(sendChannel)
}

func receiveNewMatrix(receiveChannel chan map[string]control.ElevatorNode) {
	var emptyMatrix = make(map[string]control.ElevatorNode)
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		tempMatrix = <-receiveChannel
		elevatorMatrixMutex.Lock()
		if !reflect.DeepEqual(emptyMatrix, tempMatrix) {
			if !reflect.DeepEqual(matrixBeingHandled, tempMatrix) {
				copyMapByValue(tempMatrix, elevatorMatrix)
				copyMapByValue(tempMatrix, matrixBeingHandled)
				orderArray = createOrderArray()
				tempOrder := tempMatrix[control.LocalAddress]
				Save(backupOrderFilePath, tempOrder)
				//Load(backupOrderFilePath, tempOrder)
				//fmt.Println("Printing on receive thread")
				//fmt.Println(tempOrder)
				//Check(err)
			}
		}
		if receivedFirstMatrix == false {
			receivedFirstMatrix = true
		}
		elevatorMatrixMutex.Unlock()
	}
}

func sendNewMatrix(sendChannel chan map[string]control.ElevatorNode) {
	var emptyMatrix = make(map[string]control.ElevatorNode)
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		elevatorMatrixMutex.Lock()
		if openSendChan {
			copyMapByValue(elevatorMatrix, tempMatrix)
			if !reflect.DeepEqual(emptyMatrix, tempMatrix) {
				if !reflect.DeepEqual(matrixBeingHandled, tempMatrix) {
					//fmt.Println("A completed order was sent")
					sendChannel <- tempMatrix
					tempOrder := tempMatrix[control.LocalAddress]
					Save(backupOrderFilePath, tempOrder)
					//Load(backupOrderFilePath, tempOrder)
					//fmt.Println("Printing on send thread")
					//fmt.Println(tempOrder)
					//Check(err)
					copyMapByValue(tempMatrix, matrixBeingHandled)
				}
			}
			openSendChan = false
		}
		elevatorMatrixMutex.Unlock()
	}
}

//Utility functions
//Put these in another module to promote code maintainability

func copyMapByValue(originalMap map[string]control.ElevatorNode, newMap map[string]control.ElevatorNode) {
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

///////////////////////////////////////////////////////////////

func Run(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(3)
	elevatorModuleInit()

	go lightThread()
	go elevatorMovementThread()
	go communicationThread(sendChannel, receiveChannel)
	wg.Wait()
}
