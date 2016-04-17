package elevator

import(

)

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
