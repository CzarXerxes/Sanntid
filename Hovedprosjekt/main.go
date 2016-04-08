package main

import (
	//"control"
	//"driver"
	//"elevator"
	"fmt"
	"runtime"
	//"user"
	"network"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	done := make(chan bool)
	/*
		UserToControlChan := make(chan user.ElevatorOrder)
		ElevatorToControlChan := make(chan map[int]control.ElevatorNode)
		ControlToElevatorChan := make(chan map[int]control.ElevatorNode)

		go control.Run(ControlToElevatorChan, ElevatorToControlChan, UserToControlChan)
		go elevator.Run(ElevatorToControlChan, ControlToElevatorChan)
		go user.Run(UserToControlChan)
	*/
	go network.Run()
	<-done
	fmt.Println("Done")
}
