package main

import (
	"driver"
	//"fmt"
)

const (
	Down  = -1
	Still = 0
	Up    = 1
)

func elevatorModuleInit() {
	driver.Io_init()
}

func setDirection(direction int) {
	driver.Io_set_bit(driver.MOTOR)
	driver.Io_write_analog(driver.MOTORDIR, direction*driver.MotorSpeed)
}

func main() {
	elevatorModuleInit()
	driver.Io_set_bit(driver.BUTTON_DOWN2)
	setDirection(Up)
}
