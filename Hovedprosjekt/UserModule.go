package main

import (
	"driver"
	"fmt"
)




type ElevatorOrder struct {
	orderType driver.Elev_button_type_t //Down order = -1 Up order = 1, Internal order = 0
	floor     int //1 indexed(Floor 1 = 1, Floor 2 = 2 ...)
	direction bool //Up = True, down = false
}

func init_userIO() {
	driver.Elev_init()
}

func check_buttons()

func receive_order() ElevatorOrder {
	for(i=0; i<driver.N_FLOORS; i++){
		for(j=0; j<N_BUTTONS; j++){
			if Elev_get_button_signal(button Elev_button_type_t, i)
}

func main() {
	init_userIO()
	for {
		newOrder := receive_order()
	}
}
