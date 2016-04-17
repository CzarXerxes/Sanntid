package elevator

import(
	"control"
	"time"
	"driver"
	"reflect"
)

var matrixBeingHandled map[string]control.ElevatorNode

func communicationThread(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	go receiveNewMatrix(receiveChannel)
	go sendNewMatrix(sendChannel)
}

func receiveNewMatrix(receiveChannel chan map[string]control.ElevatorNode) {
	var emptyMatrix = make(map[string]control.ElevatorNode)
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		tempMatrix = <-receiveChannel
		elevatorMatrixMutex.Lock()
		if !reflect.DeepEqual(emptyMatrix, tempMatrix) {
			if !reflect.DeepEqual(matrixBeingHandled, tempMatrix) {
				copyMapByValue(tempMatrix, elevatorMatrix)
				copyMapByValue(tempMatrix, matrixBeingHandled)
				orderArray = createOrderArray()
				tempOrder := tempMatrix[control.LocalAddress]
				Save(backupOrderFilePath, tempOrder)
			}
		}
		if receivedFirstMatrix == false {
			receivedFirstMatrix = true
		}
		elevatorMatrixMutex.Unlock()
	}
}

func sendNewMatrix(sendChannel chan map[string]control.ElevatorNode) {
	var emptyMatrix = make(map[string]control.ElevatorNode)
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		elevatorMatrixMutex.Lock()
		if openSendChan {
			copyMapByValue(elevatorMatrix, tempMatrix)
			if !reflect.DeepEqual(emptyMatrix, tempMatrix) {
				if !reflect.DeepEqual(matrixBeingHandled, tempMatrix) {
					sendChannel <- tempMatrix
					tempOrder := tempMatrix[control.LocalAddress]
					Save(backupOrderFilePath, tempOrder)
					copyMapByValue(tempMatrix, matrixBeingHandled)
				}
			}
			openSendChan = false
		}
		elevatorMatrixMutex.Unlock()
	}
}
