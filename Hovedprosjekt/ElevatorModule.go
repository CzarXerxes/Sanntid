package main

import (
	"driver"
	"fmt"
	"time"
	"sync"
)

//See elev.go for enum declarations for use with elev functions

var state int
const(
	Downward = -1
	Still = 0
	Upward = 1
)

var currentFloor int
var orderArray [driver.N_FLOORS]int//0 = Do not stop, 1 = Stop
var lightArray [driver.N_BUTTONS][driver.N_FLOORS]int//0 = Do not turn on light; 1 = Turn on light



func elevatorModuleInit() {
	for i := 0; i < driver.N_BUTTONS; i++{
		for j := 0; j < driver.N_FLOORS; j++{
			lightArray[i][j] = 0
		}
	}
	driver.Elev_init()
	for(getCurrentFloor() == -1){
		setDirection(driver.DIRN_DOWN)
	}
	setDirection(driver.DIRN_STOP)
	driver.Elev_set_floor_indicator(getCurrentFloor())
	state = Still
	fmt.Println(getCurrentFloor())
}

func setDirection(direction driver.Elev_motor_direction_t) {
	driver.Elev_set_motor_direction(direction)
}

func getCurrentFloor() int{	
	return driver.Elev_get_floor_sensor_signal()
}


func setLights(lightArray [driver.N_BUTTONS][driver.N_FLOORS]int){
	for i := 0; i < driver.N_BUTTONS; i++{
		for j := 0; j < driver.N_FLOORS; j++{
			driver.Elev_set_button_lamp(driver.Elev_button_type_t(i), j, lightArray[i][j])	
		}	
	}
}

func getOrderArray() [driver.N_FLOORS]int{//Implement differently. Currently just test
	var tempArray [driver.N_FLOORS]int
	for j := 0; j < driver.N_FLOORS; j++{
		tempArray[j] = 0
	}
	return tempArray
}

func getLightArray() [driver.N_BUTTONS][driver.N_FLOORS]int{//Implement differently. Currently just test
	var tempArray [driver.N_BUTTONS][driver.N_FLOORS]int
	for i := 0; i < driver.N_BUTTONS; i++{
		for j := 0; j < driver.N_FLOORS; j++{
			tempArray[i][j] = 0
		}
	}
	return tempArray
}

func noPendingOrders() bool{
	for i := 0; i < driver.N_FLOORS; i++{
		if(getOrderArray()[i] != 0){
			return false
		}
	}
	return true
}

func calculateState(state int) int{//Finds new state(Upward,Downward or Still) based on current state and pending orders
	if(noPendingOrders()){
		return Still
	}
	switch state {
	case Still:
		for i := 0; i < driver.N_FLOORS; i++{
			if(getOrderArray()[i] != 0){
				if(i == getCurrentFloor()){
					return Still
				}else if(i < getCurrentFloor()){
					return Downward
				}else if(i > getCurrentFloor()){
					return Upward
				}
			}
		}	 
	case Upward:
		for i := getCurrentFloor(); i < driver.N_FLOORS; i++{
			if(orderArray[i] == 1){
				return Upward
			}else{
				return Downward
			}
		}		
	case Downward:
		for i:= 0; i < getCurrentFloor(); i++{
			if(orderArray[i] == 1){
				return Downward
			}else{
				return Upward
			}
		}
	}
	return Still
}

func stopElevator(){//Stop elevator, open doors for 5 sec
	setDirection(driver.DIRN_STOP)
	driver.Elev_set_door_open_lamp(1)
	time.Sleep(time.Second * 5)
	driver.Elev_set_door_open_lamp(0)
}

func lightThread(){
	for{
		setLights(getLightArray())
	}
}

func elevatorMovementThread(){
	for{
		switch state {
		case Still:
			if(getOrderArray()[getCurrentFloor()] != 0){		
				stopElevator()
			}
			state = calculateState(Still)
			
			
		case Downward:
			for getCurrentFloor() != -1{
				setDirection(driver.DIRN_DOWN)
			}
			for getCurrentFloor() == -1{//OBS Kanskje det finnes en mer intelligent måte å gjøre dette på
			}
			if (getOrderArray()[getCurrentFloor()] == 1){
				stopElevator()
				state = calculateState(Downward)
			}
		case Upward:
			for getCurrentFloor() != -1{
				setDirection(driver.DIRN_UP)
			}
			for getCurrentFloor() == -1{//OBS Kanskje det finnes en mer intelligent måte å gjøre dette på
			}
			if (getOrderArray()[getCurrentFloor()] == 1){
				stopElevator()
				state = calculateState(Upward)
			}
		default:
			setDirection(driver.DIRN_STOP)
		}
	}
}
		
func main(){
	wg := new(sync.WaitGroup)
	wg.Add(2)
	
	elevatorModuleInit()
	
	
	go lightThread()
	go elevatorMovementThread()
	wg.Wait()
}
