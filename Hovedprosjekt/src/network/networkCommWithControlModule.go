package network

import(
	"control"
	"reflect"
	"time"
)


//In itialization phase, the elevators address(IP address if in online mode) needs to be sent to the control module
func sendInitialAddressToControlModule(address string, initializeAddressChannel chan string) {
	initializeAddressChannel <- address
}

func communicateWithElevatorThread(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	go receiveFromElevatorThread(receiveChannel)
	go sendToElevatorThread(sendChannel)
}

func receiveFromElevatorThread(receiveChannel chan map[string]control.ElevatorNode) {
	var tempOrderMap = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if !sendOrderMapToElevator {
			tempOrderMap = <-receiveChannel
			if !reflect.DeepEqual(orderMapInTransit, tempOrderMap) {
				elevatorOrderMapMutex.Lock()
				control.CopyMapByValue(tempOrderMap, orderMapInTransit)
				elevatorOrderMapMutex.Unlock()
				sendOrderMapToRouter = true
			}
		}
	}
}

func sendToElevatorThread(sendChannel chan map[string]control.ElevatorNode) {
	var tempOrderMap = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		if sendOrderMapToElevator {
			elevatorOrderMapMutex.Lock()
			control.CopyMapByValue(orderMapInTransit, tempOrderMap)
			elevatorOrderMapMutex.Unlock()
			sendChannel <- tempOrderMap
			sendOrderMapToElevator = false
		}
	}
}
