package elevator

import(
	"control"
	"time"
	"driver"
	"reflect"
)

var orderMapBeingHandled map[string]control.ElevatorNode

func communicationWithControlThread(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	go receiveNewOrderMap(receiveChannel)
	go sendNewOrderMap(sendChannel)
}

func receiveNewOrderMap(receiveChannel chan map[string]control.ElevatorNode) {
	var emptyOrderMap = make(map[string]control.ElevatorNode)
	var tempOrderMap = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		tempOrderMap = <-receiveChannel
		elevatorOrderMapMutex.Lock()
		if !reflect.DeepEqual(emptyOrderMap, tempOrderMap) {
			if !reflect.DeepEqual(orderMapBeingHandled, tempOrderMap) {
				control.CopyMapByValue(tempOrderMap, elevatorOrderMap)
				control.CopyMapByValue(tempOrderMap, orderMapBeingHandled)
				orderArray = createOrderArray()
				tempOrder := tempOrderMap[control.LocalAddress]
				driver.Save(driver.BackupOrderFilePath, tempOrder)
			}
		}
		if receivedFirstOrderMap == false {
			receivedFirstOrderMap = true
		}
		elevatorOrderMapMutex.Unlock()
	}
}

func sendNewOrderMap(sendChannel chan map[string]control.ElevatorNode) {
	var emptyOrderMap = make(map[string]control.ElevatorNode)
	var tempOrderMap = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		elevatorOrderMapMutex.Lock()
		if openSendChan {
			control.CopyMapByValue(elevatorOrderMap, tempOrderMap)
			if !reflect.DeepEqual(emptyOrderMap, tempOrderMap) {
				if !reflect.DeepEqual(orderMapBeingHandled, tempOrderMap) {
					sendChannel <- tempOrderMap
					tempOrder := tempOrderMap[control.LocalAddress]
					driver.Save(driver.BackupOrderFilePath, tempOrder)
					control.CopyMapByValue(tempOrderMap, orderMapBeingHandled)
				}
			}
			openSendChan = false
		}
		elevatorOrderMapMutex.Unlock()
	}
}
