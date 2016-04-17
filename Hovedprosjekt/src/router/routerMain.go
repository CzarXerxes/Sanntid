package router

import (
	"control"
	"sync"
	"driver"
	"net"
	"strings"
)

var routerIPAddress string

var elevatorWhichSentTheOrderMutex = &sync.Mutex{}
var connectionMutex = &sync.Mutex{}

func getRouterIP() string{
	return driver.IP
}

func routerModuleInit() {
	routerIPAddress = getRouterIP()
	portString := []string{":", driver.Port}
	elevatorSocketListener, _ = net.Listen("tcp", strings.Join(portString, ""))
}


func Run(){
	elevatorChannel := make(chan map[string]control.ElevatorNode)
	wg := new(sync.WaitGroup)
	wg.Add(5)
	routerModuleInit()
	go connectNewElevatorsThread(wg, elevatorChannel)
	go checkElevatorStillConnectedThread()
	go tellElevatorStillConnectedThread()
	go getOrderMapThread(elevatorChannel)
	go sendOrderMapThread()
	wg.Wait()
}
