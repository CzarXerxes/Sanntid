package main

import (
	"control"
	"driver"
	"elevator"
	"fmt"
	"runtime"
	//"user"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	done := make(chan bool)

	//UserControlChan := make(chan user.ElevatorOrder)
	ElevatorToControlChan := make(chan map[int]control.ElevatorNode)
	ControlToElevatorChan := make(chan map[int]control.ElevatorNode)

	go control.Run(ControlToElevatorChan, ElevatorToControlChan)
	go elevator.Run(ElevatorToControlChan, ControlToElevatorChan)
	//go user.Run(UserControlChan)

	<-done
	fmt.Println("Done")
}
