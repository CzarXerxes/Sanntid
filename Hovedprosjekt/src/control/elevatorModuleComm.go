package control

import (
	"driver"
	"sync"
	"time"
)


var openSendChanElevator bool = false

func getElevatorState() ElevatorNode {
	var elevator ElevatorNode
	elevator.CurrentFloor = 1
	elevator.CurrentDirection = Still
	var temp [driver.N_BUTTONS][driver.N_FLOORS]bool
	elevator.CurrentOrders = temp
	return elevator
}

func ordersEmpty(elevator ElevatorNode) bool {
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			if elevator.CurrentOrders[i][j] {
				return false
			}
		}
	}
	return true
}

func elevatorThread(sendChannel chan map[string]ElevatorNode, receiveChannel chan map[string]ElevatorNode) {
	go receiveNewMatrixElevator(receiveChannel)
	go sendNewMatrixElevator(sendChannel)
}

func receiveNewMatrixElevator(receiveChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		tempMatrix := <-receiveChannel
		elevatorMatrixMutex.Lock()
		copyMapByValue(tempMatrix, elevatorMatrix)
		elevatorMatrixMutex.Unlock()
		if !elevatorIsOffline {
			openSendChanNetwork = true
		}
	}
}

func sendNewMatrixElevator(sendChannel chan map[string]ElevatorNode) {
	var tempMatrix = make(map[string]ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanElevator {
			elevatorMatrixMutex.Lock()
			copyMapByValue(elevatorMatrix, tempMatrix)
			elevatorMatrixMutex.Unlock()
			sendChannel <- tempMatrix
			if !elevatorIsOffline {
				openSendChanElevator = false
			}
		}
	}
}
