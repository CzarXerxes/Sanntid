package router

import (
	"control"
	"encoding/gob"
	//"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"reflect"
	"sync"
	"time"
)




var elevatorListener net.Listener

var elevatorAliveConnections = make(map[string]net.Conn) //Dictionary with ipAddress:connectionSocket
var elevatorCommConnections = make(map[string]net.Conn)



func connectNewElevatorsThread(wg *sync.WaitGroup, channel chan map[string]control.ElevatorNode) {
	for {
		time.Sleep(time.Millisecond * 10)
		aliveConnection, err := elevatorListener.Accept()
		if err != nil {
			panic(err)
		}
		commConnection, err := elevatorListener.Accept()
		if err != nil {
			panic(err)
		}
		elevatorIPAddress := aliveConnection.RemoteAddr().String()
		elevatorAliveConnections[elevatorIPAddress] = aliveConnection
		elevatorCommConnections[elevatorIPAddress] = commConnection

		//aliveConnection.SetReadDeadline(time.Now().Add(2 * time.Second))
		//commConnection.SetReadDeadline(time.Now().Add(2 * time.Second))

		elevatorEncoders[elevatorIPAddress] = gob.NewEncoder(commConnection)
		elevatorDecoders[elevatorIPAddress] = gob.NewDecoder(commConnection)

		go receiveIncoming(elevatorDecoders[elevatorIPAddress], channel)
		wg.Add(1)

		var tempMatrix = make(map[string]control.ElevatorNode)
		elevatorDecoders[elevatorIPAddress].Decode(&tempMatrix)
		connectionMutex.Lock()
		initialNode := tempMatrix[elevatorIPAddress]
		matrixInTransit[elevatorIPAddress] = initialNode
		connectionMutex.Unlock()
		for elevator, _ := range elevatorAliveConnections {
			elevatorEncoders[elevator].Encode(matrixInTransit)
		}
	}
}



func tellElevatorStillConnected(elevatorIP string) bool {
	text := "Still alive"
	if elevatorAliveConnections[elevatorIP] == nil {
		fmt.Println("Connection failed because there is no connection")
		return false
	}
	_, err := fmt.Fprintf(elevatorAliveConnections[elevatorIP], text)
	if err != nil {
		fmt.Println("Failed because there was a write error to the socket")
		return false
	}
	return true
}

func tellElevatorStillConnectedThread() {
	for {
		time.Sleep(time.Millisecond * 500)
		for elevator, _ := range elevatorAliveConnections {
			if !tellElevatorStillConnected(elevator) {
				elevatorIsDead(elevator)
			}
		}
	}
}



func checkElevatorStillConnected(elevatorIP string) bool {
	buf := make([]byte, 1024)
	if elevatorAliveConnections[elevatorIP] == nil {
		fmt.Println("Connection failed because there is no connection")
		return false
	}
	_, err := elevatorAliveConnections[elevatorIP].Read(buf)
	if err != nil {
		fmt.Println("Failed because there was a read error to the socket")
		return false
	}
	//fmt.Printf("Message received :: %s\n", string(buf[:n]))
	return true
}

func checkElevatorStillConnectedThread() {
	for {
		time.Sleep(time.Millisecond * 500)
		for elevator, _ := range elevatorAliveConnections {
			if !checkElevatorStillConnected(elevator) {
				elevatorIsDead(elevator)
			}
		}
	}
}

func elevatorIsDead(elevator string) {
	elevatorAliveConnections[elevator].Close()
	elevatorCommConnections[elevator].Close()
	time.Sleep(time.Second * 1)
	delete(elevatorAliveConnections, elevator)
	delete(elevatorCommConnections, elevator)
	delete(elevatorEncoders, elevator)
	delete(elevatorDecoders, elevator)
	delete(matrixInTransit, elevator)
	for elevator, _ := range elevatorAliveConnections {
		elevatorEncoders[elevator].Encode(matrixInTransit)
	}
}
