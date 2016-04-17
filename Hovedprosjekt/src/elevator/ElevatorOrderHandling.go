package elevator

import(

)

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

func getOrderArray(directionIndex int, floor int) bool { //directionIndex valid values {UpIndex, DownIndex}
	return orderArray[directionIndex][floor]
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
