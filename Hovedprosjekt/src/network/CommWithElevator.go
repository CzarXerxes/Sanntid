package network

import(
	"driver"
	"control"
	"reflect"
	"time"
	"sync"
)

func sendInitialAddressToElevator(address string, initializeAddressChannel chan string) {
	initializeAddressChannel <- address
}

func communicateWithElevatorThread(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	go receiveFromElevatorThread(receiveChannel)
	go sendToElevatorThread(sendChannel)
}

func receiveFromElevatorThread(receiveChannel chan map[string]control.ElevatorNode) {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if !sendMatrixToElevator {
			tempMatrix = <-receiveChannel
			if !reflect.DeepEqual(matrixInTransit, tempMatrix) {
				elevatorMatrixMutex.Lock()
				control.CopyMapByValue(tempMatrix, matrixInTransit)
				elevatorMatrixMutex.Unlock()
				sendMatrixToRouter = true
			}
		}
	}
}

func sendToElevatorThread(sendChannel chan map[string]control.ElevatorNode) {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if sendMatrixToElevator {
			elevatorMatrixMutex.Lock()
			control.CopyMapByValue(matrixInTransit, tempMatrix)
			elevatorMatrixMutex.Unlock()
			sendChannel <- tempMatrix
			sendMatrixToElevator = false
		}
	}
}
