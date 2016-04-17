package router

import (
	"control"
	"sync"
	"driver"
	"net"
	"strings"
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
}


func Run(){
	elevatorChannel := make(chan map[string]control.ElevatorNode)
	wg := new(sync.WaitGroup)
	wg.Add(5)
	routerModuleInit()
	go connectNewElevatorsThread(wg, elevatorChannel)
	go checkElevatorStillConnectedThread()
	go tellElevatorStillConnectedThread()
	go getMatrixThread(elevatorChannel)
	go sendMatrixThread()
	wg.Wait()
}
