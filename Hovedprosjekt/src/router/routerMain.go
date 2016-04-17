package router

import (
	"control"
	"sync"
)

var routerIPAddress string
var sendMatrix bool

var elevatorWhichSentTheOrderMutex = &sync.Mutex{}
var connectionMutex = &sync.Mutex{}

func getRouterIP() {
	routerIPAddress = driver.IP
}

func routerModuleInit() {
	getRouterIP()
	portString := []string{":", driver.Port}
	elevatorListener, _ = net.Listen("tcp", strings.Join(portString, ""))
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
