package router

import (
	"control"
	"encoding/gob"
	"reflect"
	"time"
)

var elevatorEncoders = make(map[string]*gob.Encoder)
var elevatorDecoders = make(map[string]*gob.Decoder)

var matrixInTransit = make(map[string]control.ElevatorNode)
var elevatorWhichSentTheOrder string

func receiveIncoming(dec *gob.Decoder, channel chan map[string]control.ElevatorNode) {
	var newMap = make(map[string]control.ElevatorNode)
	for {
		dec.Decode(&newMap)
		channel <- newMap
	}
}

func getMatrixThread(channel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		tempMatrix := <-channel
		if !reflect.DeepEqual(matrixInTransit, tempMatrix) {
			connectionMutex.Lock()
			control.CopyMapByValue(tempMatrix, matrixInTransit)
			connectionMutex.Unlock()
			sendMatrix = true
		}
	}
}

func sendMatrixThread() {
	var tempMatrix = make(map[string]control.ElevatorNode)
	for {
		time.Sleep(time.Millisecond * 10)
		connectionMutex.Lock()
		control.CopyMapByValue(matrixInTransit, tempMatrix)
		connectionMutex.Unlock()
		if sendMatrix {
			for elevator, _ := range elevatorAliveConnections {
				elevatorEncoders[elevator].Encode(tempMatrix)
			}
		}
		sendMatrix = false
	}
}
