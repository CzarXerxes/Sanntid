package elevator

import(
)

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
