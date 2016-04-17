package router

import (
	"control"
	"encoding/gob"
	"reflect"
	"time"
)

var elevatorEncoders = make(map[string]*gob.Encoder)
var elevatorDecoders = make(map[string]*gob.Decoder)

var orderMapInTransit = make(map[string]control.ElevatorNode)
var elevatorWhichSentTheOrder string
var shouldSendOrderMap bool

func receiveNewElevatorStatus(dec *gob.Decoder, channel chan map[string]control.ElevatorNode) {
	var newMap = make(map[string]control.ElevatorNode)
	for {
		dec.Decode(&newMap)
		channel <- newMap
	}
}

func getOrderMapThread(channel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		tempOrderMap := <-channel
		if !reflect.DeepEqual(orderMapInTransit, tempOrderMap) {
			connectionMutex.Lock()
			control.CopyMapByValue(tempOrderMap, orderMapInTransit)
			connectionMutex.Unlock()
			shouldSendOrderMap = true
		}
	}
}

func sendOrderMapThread() {
	var tempOrderMap = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		connectionMutex.Lock()
		control.CopyMapByValue(orderMapInTransit, tempOrderMap)
		connectionMutex.Unlock()
		if shouldSendOrderMap {
			for elevator, _ := range elevatorAliveConnectionsMap {
				elevatorEncoders[elevator].Encode(tempOrderMap)
			}
		}
		shouldSendOrderMap = false
	}
}
