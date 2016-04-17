package elevator

import(
	"driver"
	"control"
	"time"
)

var currentFloor int
var elevatorIsMoving bool = false
var receivedFirstOrderMap bool = false

func getCurrentFloor() int {
	return driver.Elev_get_floor_sensor_signal()
}

func updateCurrentDirection(direction driver.Elev_motor_direction_t) {
	elevatorOrderMapMutex.Lock()
	var tempOrderMap = make(map[string]control.ElevatorNode)
	control.CopyMapByValue(elevatorOrderMap, tempOrderMap)
	tempNode := tempOrderMap[control.LocalAddress]
	tempNode.CurrentDirection = direction
	tempOrderMap[control.LocalAddress] = tempNode
	control.CopyMapByValue(tempOrderMap, elevatorOrderMap)
	elevatorOrderMapMutex.Unlock()
}

func updateCurrentFloor(floor int) {
	elevatorOrderMapMutex.Lock()
	var tempOrderMap = make(map[string]control.ElevatorNode)
	control.CopyMapByValue(elevatorOrderMap, tempOrderMap)
	tempNode := tempOrderMap[control.LocalAddress]
	tempNode.CurrentFloor = floor
	tempOrderMap[control.LocalAddress] = tempNode
	control.CopyMapByValue(tempOrderMap, elevatorOrderMap)
	elevatorOrderMapMutex.Unlock()
}

func calculateNewDirection() int { 
	if noPendingOrders() {
		return Still
	}

	switch currentDirection {
	case Still:
		for i := 0; i < driver.N_FLOORS; i++ {
			if elevatorShouldStop(UpIndex, i) || elevatorShouldStop(DownIndex, i) {
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
			if elevatorShouldStop(UpIndex, i) {
				return Upward
			}
		}
		for i := 0; i < currentFloor; i++ {
			if elevatorShouldStop(DownIndex, i) || elevatorShouldStop(UpIndex, i) {
				return Downward
			}
		}
	case Downward:
		for i := 0; i <= currentFloor; i++ {
			if elevatorShouldStop(DownIndex, i) {
				return Downward
			}
		}
		for i := currentFloor + 1; i < driver.N_FLOORS; i++ {
			if elevatorShouldStop(DownIndex, i) || elevatorShouldStop(UpIndex, i) {
				return Upward
			}
		}
	}
	return Still
}

func setElevatorDirection(direction driver.Elev_motor_direction_t) {
	driver.Elev_set_motor_direction(direction)
	if direction != driver.DIRN_STOP {
		elevatorIsMoving = true
	} else {
		elevatorIsMoving = false
	}
}

func startElevator(direction driver.Elev_motor_direction_t) {
	for getCurrentFloor() != -1 {
		setElevatorDirection(direction)
		time.Sleep(time.Millisecond * 10)
	}
	for getCurrentFloor() == -1 {
		time.Sleep(time.Millisecond * 10)
	}
	tempFloor := currentFloor
	currentFloor = getCurrentFloor()
	if tempFloor != currentFloor {
		updateCurrentFloor(currentFloor)
	}
}

func stopElevator() {
	setElevatorDirection(driver.DIRN_STOP)
	driver.Elev_set_door_open_lamp(1)
	time.Sleep(time.Second * 3)
	driver.Elev_set_door_open_lamp(0)
}

func floorIsReached() {
	stopElevator()
	deleteOrders()
	currentDirection = calculateNewDirection()
	deleteOrders()
	updateCurrentDirection(driver.Elev_motor_direction_t(currentDirection))
}

func elevatorMovementThread() {
	for {
		time.Sleep(time.Millisecond * 10)
		if receivedFirstOrderMap {
			switch currentDirection {
			case Still:
				updateCurrentDirection(driver.Elev_motor_direction_t(currentDirection))
				if elevatorShouldStop(UpIndex, currentFloor) || elevatorShouldStop(DownIndex, currentFloor) {
					floorIsReached()
				}
				currentDirection = calculateNewDirection()
				deleteOrders()
			case Downward:
				updateCurrentDirection(driver.Elev_motor_direction_t(currentDirection))
				if noPendingOrdersInDirection(DownIndex) {
					if elevatorShouldStop(UpIndex, currentFloor) || elevatorShouldStop(DownIndex, currentFloor) {
						floorIsReached()
					}
				} else {
					if elevatorShouldStop(DownIndex, currentFloor) {
						floorIsReached()
					}
				}
				if currentDirection == Downward {
					startElevator(driver.DIRN_DOWN)
				}
			case Upward:
				updateCurrentDirection(driver.Elev_motor_direction_t(currentDirection))
				if noPendingOrdersInDirection(UpIndex) {
					if elevatorShouldStop(UpIndex, currentFloor) || elevatorShouldStop(DownIndex, currentFloor) {
						floorIsReached()
					}
				} else {
					if elevatorShouldStop(UpIndex, currentFloor) {
						floorIsReached()
					}
				}
				if currentDirection == Upward {
					startElevator(driver.DIRN_UP)
				}
			default:
				setElevatorDirection(driver.DIRN_STOP)
			}
		}
	}
}
