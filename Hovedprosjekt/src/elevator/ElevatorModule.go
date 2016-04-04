package elevator

import (
	"control"
	"driver"
	"fmt"
	"sync"
	"time"
)

//See elev.go for enum declarations for use with elev functions

var state int
var currentFloor int
var openSendChan bool = false
var elevatorMatrix map[int]control.ElevatorNode

const (
	Downward = -1
	Still    = 0
	Upward   = 1
)

//Extend orderArray to have seperate columns for stopping upwards and downwards
var orderArray [driver.N_FLOORS]int                   //0 = Do not stop, 1 = Stop
var lightArray [driver.N_BUTTONS][driver.N_FLOORS]int //0 = Do not turn on light; 1 = Turn on light

//Initialization function
func elevatorModuleInit() {
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			lightArray[i][j] = 0
		}
	}
	driver.Elev_init()
	for getCurrentFloor() == -1 {
		setDirection(driver.DIRN_DOWN)
	}
	setDirection(driver.DIRN_STOP)
	driver.Elev_set_floor_indicator(getCurrentFloor())
	state = Still
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
func getOrderArray() [2][driver.N_FLOORS]bool {
	var tempArray [2][driver.N_FLOORS]bool//tempArray[0][driver.N_FLOORS] corresponds to orders to complete on way UP, tempArray[1][driver.N_FLOORS] corresponds to orders to complete on way DOWN
	var tempNode control.ElevatorNode
	tempNode = elevatorMatrix[control.LocalAddress]
	for i := 0; i < 2; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			tempArray[i][j] = tempNode.CurrentOrders[i][j]
		}
	}
	for i:= tempNode.CurrentFloor; i <driver.N_FLOORS; i++{
		tempArray[0][i] = tempNode.CurrentOrders[2][i] || tempArray[0][i]
	}
	for i:= 0; i < tempNode.CurrentFloor; i++{
		tempArray[1][i] = tempNode.CurrentOrders[2][i] || tempArray[1][i]
}

func noPendingOrders() bool {
	for i := 0; i < driver.N_FLOORS; i++ {
		if getOrderArray(Downward)[i] != 0 && getOrderArray(Upward)[i] != 0 {
			return false
		}
	}
	return true
}

func calculateState(state int) int { //Finds new state(Upward,Downward or Still) based on current state and pending orders
	if noPendingOrders() {
		return Still
	}
	switch state {
	case Still:
		for i := 0; i < driver.N_FLOORS; i++ {
			if getOrderArray(Still)[i] != 0 {
				if i == getCurrentFloor() {
					return Still
				} else if i < getCurrentFloor() {
					return Downward
				} else if i > getCurrentFloor() {
					return Upward
				}
			}
		}
	case Upward:
		for i := getCurrentFloor(); i < driver.N_FLOORS; i++ {
			if orderArray[i] == 1 {
				return Upward
			}
		}
	case Downward:
		for i := 0; i < getCurrentFloor(); i++ {
			if orderArray[i] == 1 {
				return Downward
			}
		}
	}
	return Still
}

//Elevator movement functions
func setDirection(direction driver.Elev_motor_direction_t) {
	driver.Elev_set_motor_direction(direction)
}

func stopElevator() { //Stop elevator, open doors for 5 sec
	setDirection(driver.DIRN_STOP)
	driver.Elev_set_door_open_lamp(1)
	time.Sleep(time.Second * 5)
	driver.Elev_set_door_open_lamp(0)
}

//Main threads
func lightThread() {
	for {
		setLights(getLightArray())
	}
}

func elevatorMovementThread() {
	for {
		switch state {
		case Still:
			if getOrderArray(Still)[getCurrentFloor()] != 0 {
				stopElevator()
			}
			state = calculateState(Still)

		case Downward:
			for getCurrentFloor() != -1 {
				setDirection(driver.DIRN_DOWN)
			}
			for getCurrentFloor() == -1 { //OBS Kanskje det finnes en mer intelligent måte å gjøre dette på
			}
			if getOrderArray(Downward)[getCurrentFloor()] == 1 {
				stopElevator()
				state = calculateState(Downward)
			}
		case Upward:
			for getCurrentFloor() != -1 {
				setDirection(driver.DIRN_UP)
			}
			for getCurrentFloor() == -1 { //OBS Kanskje det finnes en mer intelligent måte å gjøre dette på
			}
			if getOrderArray(Upward)[getCurrentFloor()] == 1 {
				stopElevator()
				state = calculateState(Upward)
			}
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
		elevatorMatrix := <-receiveChannel
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
