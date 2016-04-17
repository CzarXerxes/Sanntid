package control

import (
	"driver"
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
	go receiveNewOrderMapElevator(receiveChannel)
	go sendNewOrderMapElevator(sendChannel)
}

func receiveNewOrderMapElevator(receiveChannel chan map[string]ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		tempOrderMap := <-receiveChannel
		elevatorOrderMapMutex.Lock()
		CopyMapByValue(tempOrderMap, elevatorOrderMap)
		elevatorOrderMapMutex.Unlock()
		if !elevatorIsOffline {
			openSendChanNetwork = true
		}
	}
}

func sendNewOrderMapElevator(sendChannel chan map[string]ElevatorNode) {
	var tempOrderMap = make(map[string]ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if openSendChanElevator {
			elevatorOrderMapMutex.Lock()
			CopyMapByValue(elevatorOrderMap, tempOrderMap)
			elevatorOrderMapMutex.Unlock()
			sendChannel <- tempOrderMap
			if !elevatorIsOffline {
				openSendChanElevator = false
			}
		}
	}
}
