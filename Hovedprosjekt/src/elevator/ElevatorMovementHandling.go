package elevator

import(
	"driver"
	"control"
	"time"
	"reflect"
)

var currentFloor int
var isMoving bool = false
var receivedFirstMatrix bool = false

func getCurrentFloor() int {
	return driver.Elev_get_floor_sensor_signal()
}

func setElevatorMatrixDirection(direction driver.Elev_motor_direction_t) {
	elevatorMatrixMutex.Lock()
	var tempMatrix = make(map[string]control.ElevatorNode)
	control.CopyMapByValue(elevatorMatrix, tempMatrix)
	tempNode := tempMatrix[control.LocalAddress]
	tempNode.CurrentDirection = direction
	tempMatrix[control.LocalAddress] = tempNode
	control.CopyMapByValue(tempMatrix, elevatorMatrix)
	elevatorMatrixMutex.Unlock()
}

func setElevatorMatrixFloor(floor int) {
	elevatorMatrixMutex.Lock()
	var tempMatrix = make(map[string]control.ElevatorNode)
	control.CopyMapByValue(elevatorMatrix, tempMatrix)
	tempNode := tempMatrix[control.LocalAddress]
	tempNode.CurrentFloor = floor
	tempMatrix[control.LocalAddress] = tempNode
	Control.CopyMapByValue(tempMatrix, elevatorMatrix)
	elevatorMatrixMutex.Unlock()
}

func calculateCurrentDirection() int { 
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

func stopElevator() {
	setDirection(driver.DIRN_STOP)
	driver.Elev_set_door_open_lamp(1)
	time.Sleep(time.Second * 3)
	driver.Elev_set_door_open_lamp(0)
}

func floorIsReached() {
	stopElevator()
	deleteOrders()
	currentDirection = calculateCurrentDirection()
	deleteOrders()
	setElevatorMatrixDirection(driver.Elev_motor_direction_t(currentDirection))
}

func elevatorMovementThread() {
	for {
		time.Sleep(time.Millisecond * 10)
		if receivedFirstMatrix {
			switch currentDirection {
			case Still:
				setElevatorMatrixDirection(driver.Elev_motor_direction_t(currentDirection))
				if getOrderArray(UpIndex, currentFloor) || getOrderArray(DownIndex, currentFloor) {
					floorIsReached()
				}
				currentDirection = calculateCurrentDirection()
				deleteOrders()
			case Downward:
				setElevatorMatrixDirection(driver.Elev_motor_direction_t(currentDirection))
				if noPendingOrdersDirection(DownIndex) {
					if getOrderArray(UpIndex, currentFloor) || getOrderArray(DownIndex, currentFloor) {
						floorIsReached()
					}
				} else {
					if getOrderArray(DownIndex, currentFloor) {
						floorIsReached()
					}
				}
				if currentDirection == Downward {
					moveElevator(driver.DIRN_DOWN)
				}
			case Upward:
				setElevatorMatrixDirection(driver.Elev_motor_direction_t(currentDirection))
				if noPendingOrdersDirection(UpIndex) {
					if getOrderArray(UpIndex, currentFloor) || getOrderArray(DownIndex, currentFloor) {
						floorIsReached()
					}
				} else {
					if getOrderArray(UpIndex, currentFloor) {
						floorIsReached()
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
