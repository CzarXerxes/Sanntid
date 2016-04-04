package user

import (
	"driver"
	//"fmt"
)

type ElevatorOrder struct {
	OrderType driver.Elev_button_type_t //Down order = 0 Up order = 1, Internal order = 2
	Floor     int                       //1 indexed(Floor 1 = 1, Floor 2 = 2 ...)
}

func init_userIO() {
	driver.Elev_init()
}

func receive_order() ElevatorOrder {
	var tempOrder ElevatorOrder
	var orderReceived bool = false
	for {
		for i := 0; i < driver.N_FLOORS; i++ {
			for j := 0; j < driver.N_BUTTONS; j++ {
				if driver.Elev_get_button_signal(driver.Elev_button_type_t(j), i) == 1 {
					tempOrder.OrderType = driver.Elev_button_type_t(j)
					tempOrder.Floor = i
					orderReceived = true
					break
				}
			}
		}
		if orderReceived {
			return tempOrder
		}
	}
}

func Run(c chan ElevatorOrder) {
	init_userIO()
	for {
		newOrder := receive_order()
		c <- newOrder
	}
}
