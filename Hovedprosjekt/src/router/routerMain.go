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


const IP = "129.241.187.153"

const backupPort = ":30000"
const elevatorPort = ":29000"


var routerIPAddress string
var sendMatrix bool

var elevatorWhichSentTheOrderMutex = &sync.Mutex{}
var connectionMutex = &sync.Mutex{}


func getRouterIP() { //Implement to find local IP address
	routerIPAddress = IP
}

func routerModuleInit() {
	getRouterIP()
	backupListener, _ = net.Listen("tcp", backupPort)
	elevatorListener, _ = net.Listen("tcp", elevatorPort)
	spawnBackup()
}


func Run(){
	elevatorChannel := make(chan map[string]control.ElevatorNode)

	wg := new(sync.WaitGroup)
	wg.Add(7)
	routerModuleInit()
	go connectNewElevatorsThread(wg, elevatorChannel)
	go checkElevatorStillConnectedThread()
	go tellElevatorStillConnectedThread()
	go tellBackupAliveThread()
	go spawnNewBackupThread()
	go getMatrixThread(elevatorChannel)
	go sendMatrixThread()

	wg.Wait()
}
