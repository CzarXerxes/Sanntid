package router

import (
	"control"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"
)

var elevatorSocketListener net.Listener
var elevatorAliveConnectionsMap = make(map[string]net.Conn)
var elevatorCommConnectionsMap = make(map[string]net.Conn)

func connectNewElevatorsThread(wg *sync.WaitGroup, channel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		aliveConnection, err := elevatorSocketListener.Accept()
		if err != nil {
			panic(err)
		}
		commConnection, err := elevatorSocketListener.Accept()
		if err != nil {
			panic(err)
		}
		elevatorIPAddress := aliveConnection.RemoteAddr().String()
		
		elevatorAliveConnectionsMap[elevatorIPAddress] = aliveConnection
		elevatorCommConnectionsMap[elevatorIPAddress] = commConnection

		elevatorEncoders[elevatorIPAddress] = gob.NewEncoder(commConnection)
		elevatorDecoders[elevatorIPAddress] = gob.NewDecoder(commConnection)

		go receiveNewElevatorStatus(elevatorDecoders[elevatorIPAddress], channel)
		wg.Add(1)

		var tempOrderMap = make(map[string]control.ElevatorNode)
		elevatorDecoders[elevatorIPAddress].Decode(&tempOrderMap)
		connectionMutex.Lock()
		initialNode := tempOrderMap[elevatorIPAddress]
		orderMapInTransit[elevatorIPAddress] = initialNode
		connectionMutex.Unlock()
		for elevator, _ := range elevatorAliveConnectionsMap {
			elevatorEncoders[elevator].Encode(orderMapInTransit)
		}
	}
}

func tellElevatorStillConnected(elevatorIP string) bool {
	text := "Still alive"
	if elevatorAliveConnectionsMap[elevatorIP] == nil {
		fmt.Println("Connection failed because there is no connection")
		return false
	}
	_, err := fmt.Fprintf(elevatorAliveConnectionsMap[elevatorIP], text)
	if err != nil {
		fmt.Println("Failed because there was a write error to the socket")
		return false
	}
	return true
}

func tellElevatorStillConnectedThread() {
	for {
		time.Sleep(time.Millisecond * 500)
		for elevator, _ := range elevatorAliveConnectionsMap {
			if !tellElevatorStillConnected(elevator) {
				elevatorIsDead(elevator)
			}
		}
	}
}

func checkElevatorStillConnected(elevatorIP string) bool {
	buf := make([]byte, 1024)
	if elevatorAliveConnectionsMap[elevatorIP] == nil {
		fmt.Println("Connection failed because there is no connection")
		return false
	}
	_, err := elevatorAliveConnectionsMap[elevatorIP].Read(buf)
	if err != nil {
		fmt.Println("Failed because there was a read error to the socket")
		return false
	}
	return true
}

func checkElevatorStillConnectedThread() {
	for {
		time.Sleep(time.Millisecond * 500)
		for elevator, _ := range elevatorAliveConnectionsMap {
			if !checkElevatorStillConnected(elevator) {
				elevatorIsDead(elevator)
			}
		}
	}
}

func elevatorIsDead(elevator string) {
	elevatorAliveConnectionsMap[elevator].Close()
	elevatorCommConnectionsMap[elevator].Close()
	time.Sleep(time.Second * 1)
	delete(elevatorAliveConnectionsMap, elevator)
	delete(elevatorCommConnectionsMap, elevator)
	delete(elevatorEncoders, elevator)
	delete(elevatorDecoders, elevator)
	delete(orderMapInTransit, elevator)
	for elevator, _ := range elevatorAliveConnectionsMap {
		elevatorEncoders[elevator].Encode(orderMapInTransit)
	}
}
