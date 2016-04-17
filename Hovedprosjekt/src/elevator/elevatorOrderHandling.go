package elevator

import(
	"driver"
	"control"
)

var orderArray [2][driver.N_FLOORS]bool 

//createOrderArray() reates a two dimensional array from the elevatorOrderMap to make it more convenient for the elvator to control itself
func createOrderArray() [2][driver.N_FLOORS]bool {
	var tempArray [2][driver.N_FLOORS]bool 
	var tempNode control.ElevatorNode
	var tempOrderMap = make(map[string]control.ElevatorNode)
	control.CopyMapByValue(elevatorOrderMap, tempOrderMap)

	tempNode = tempOrderMap[control.LocalAddress]
	for i := 0; i < 2; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			tempArray[i][j] = tempNode.CurrentOrders[i][j]
		}
	}
	if elevatorIsMoving {
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

func elevatorShouldStop(directionIndex int, floor int) bool { 
	return orderArray[directionIndex][floor]
}

func noPendingOrdersInDirection(directionIndex int) bool {
	for i := 0; i < driver.N_FLOORS; i++ {
		if elevatorShouldStop(directionIndex, i) {
			return false
		}
	}
	return true
}

func noPendingOrders() bool {
	return noPendingOrdersInDirection(DownIndex) && noPendingOrdersInDirection(UpIndex)
}

func setOrderArray(value bool, directionIndex int, floor int) {
	elevatorOrderMapMutex.Lock()
	var tempOrderMap = make(map[string]control.ElevatorNode)
	control.CopyMapByValue(elevatorOrderMap, tempOrderMap)

	orderArray[directionIndex][floor] = value

	var tempNode control.ElevatorNode
	tempNode = tempOrderMap[control.LocalAddress]
	tempNode.CurrentOrders[directionIndex][floor] = value
	tempNode.CurrentOrders[InternalIndex][floor] = value
	tempOrderMap[control.LocalAddress] = tempNode
	control.CopyMapByValue(tempOrderMap, elevatorOrderMap)
	openSendChan = true
	elevatorOrderMapMutex.Unlock()
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
