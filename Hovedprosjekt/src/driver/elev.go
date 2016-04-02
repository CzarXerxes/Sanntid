package driver  // where "driver" is the folder that contains io.go, io.c, io.h, channels.go, channels.c and driver.go
/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
*/
import "C"

const N_BUTTONS = C.N_BUTTONS
const N_FLOORS int = C.N_FLOORS
const M int = 3 

type Elev_motor_direction_t int
const(
	DIRN_DOWN Elev_motor_direction_t = -1
	DIRN_STOP = 0
	DIRN_UP = 1
)

type Elev_button_type_t int
const(
	BUTTON_CALL_UP = 0
	BUTTON_CALL_DOWN = 1
	BUTTON_COMMAND = 2
)

func Elev_init(){
	C.elev_init()
}

func Elev_set_motor_direction(dirn Elev_motor_direction_t){
	C.elev_set_motor_direction(C.elev_motor_direction_t(dirn))
}


func Elev_set_button_lamp(button Elev_button_type_t, floor int, value int){
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor), C.int(value))
}

func Elev_set_floor_indicator(floor int){
	C.elev_set_floor_indicator(C.int(floor))
}

func Elev_set_door_open_lamp(value int){
	C.elev_set_door_open_lamp(C.int(value))
}

func Elev_set_stop_lamp(value int){
	C.elev_set_stop_lamp(C.int(value))
}



func Elev_get_button_signal(button Elev_button_type_t, floor int) int{
	return int(C.elev_get_button_signal(C.elev_button_type_t(button), C.int(floor)))
}


func Elev_get_floor_sensor_signal() int{
	return int(C.elev_get_floor_sensor_signal())
}

func elev_get_stop_signal() int{
	return int(C.elev_get_stop_signal())
}

func elev_get_obstruction_signal() int{
	return int(C.elev_get_obstruction_signal())
}
