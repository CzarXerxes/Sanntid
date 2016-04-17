package control

import (
	"driver"
	"encoding/gob"
	//"fmt"
	"os"
	"sync"
	"time"
	"user"
)

func receiveOrder(receiveChannel chan user.ElevatorOrder) user.ElevatorOrder {
	newOrder := <-receiveChannel
	return newOrder
}

func userThread(receiveChannel chan user.ElevatorOrder) {
	for {
		time.Sleep(time.Millisecond * 10)
		newOrder := receiveOrder(receiveChannel)
		distributeOrder(LocalAddress, newOrder, elevatorMatrix)
		sendUpdatedMatrix()
	}
}
