package main

import (
	"control"
	"elevator"
	"fmt"
	"network"
	"runtime"
	"user"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	done := make(chan bool)
	blockUserChan := make(chan bool)
	blockNetworkChan := make(chan bool)
	UserToControlChan := make(chan user.ElevatorOrder)
	ElevatorToControlChan := make(chan map[string]control.ElevatorNode)
	ControlToElevatorChan := make(chan map[string]control.ElevatorNode)
	NetworkToControlChan := make(chan map[string]control.ElevatorNode)
	ControlToNetworkChan := make(chan map[string]control.ElevatorNode)
	InitializeAddressChan := make(chan string)

	go control.Run(InitializeAddressChan, blockUserChan, blockNetworkChan, ControlToNetworkChan, NetworkToControlChan, ControlToElevatorChan, ElevatorToControlChan, UserToControlChan)
	go elevator.Run(ElevatorToControlChan, ControlToElevatorChan)
	go user.Run(blockUserChan, UserToControlChan)
	go network.Run(InitializeAddressChan, blockNetworkChan, NetworkToControlChan, ControlToNetworkChan)
	<-done
	fmt.Println("Done")
}
