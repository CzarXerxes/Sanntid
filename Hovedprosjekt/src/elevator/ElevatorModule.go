package elevator

import (
	"control"
	"driver"
	"encoding/gob"
	//"fmt"
	"os"
	"reflect"
	"sync"
	"time"
)

//See elev.go for enum declarations for use with elev functions
var backupOrderFilePath = "/home/student/Desktop/Heis/backupOrders.gob"

var currentDirection int

const (
	Downward = -1
	Still    = 0
	Upward   = 1
)

var currentFloor int
var isMoving bool = false

const (
	UpIndex       = 0
	DownIndex     = 1
	InternalIndex = 2
)

var receivedFirstMatrix bool = false
var openSendChan bool = false
var elevatorMatrix map[string]control.ElevatorNode
var matrixBeingHandled map[string]control.ElevatorNode

var elevatorMatrixMutex = &sync.Mutex{}

//Extend orderArray to have seperate columns for stopping upwards and downwards
var orderArray [2][driver.N_FLOORS]bool               //false = Do not stop, true = Stop
var lightArray [driver.N_BUTTONS][driver.N_FLOORS]int //0 = Do not turn on light; 1 = Turn on light


func BoolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func IntToBool(i int) bool {
	if i == 1 {
		return true
	} else {
		return false
	}
}

func copyMapByValue(originalMap map[string]control.ElevatorNode, newMap map[string]control.ElevatorNode) {
	for k, _ := range newMap {
		delete(newMap, k)
	}
	for k, v := range originalMap {
		newMap[k] = v
	}
}

func Save(path string, object interface{}) error {
	file, err := os.Create(path)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func Load(path string, object interface{}) error {
	file, err := os.Open(path)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}

///////////////////////////////////////////////////////////////

func Run(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(3)
	elevatorModuleInit()

	go lightThread()
	go elevatorMovementThread()
	go communicationThread(sendChannel, receiveChannel)
	wg.Wait()
}
