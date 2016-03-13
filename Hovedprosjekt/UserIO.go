package main

import (
	"driver"
	"fmt"
)




type order struct {
	orderType int //Obstruction = -1, Stop = 0, External order = 1, Internal order = 2
	floor     int
	direction bool //Up = True, down = false
}

func init_userIO() {
	driver.Io_init()
}

func receive_order() order {
}

func main() {
	init_userIO()
	for {
		newOrder := receive_order()
	}
}
