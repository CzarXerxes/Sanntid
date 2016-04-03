package main

import (
	"driver"
	"fmt"
)

type ElevatorOrder struct {
	orderType driver.Elev_button_type_t //Down order = 0 Up order = 1, Internal order = 2
	floor     int                       //1 indexed(Floor 1 = 1, Floor 2 = 2 ...)
}

func init_userIO() {
	driver.Elev_init()
}

func receive_order() ElevatorOrder {
	var tempOrder ElevatorOrder
	var orderReceived bool = false
	for i := 0; i < driver.N_FLOORS; i++ {
		for j := 0; j < driver.N_BUTTONS; j++ {
			if driver.Elev_get_button_signal(driver.Elev_button_type_t(j), i) == 1 {
				tempOrder.orderType = driver.Elev_button_type_t(j)
				tempOrder.floor = i
				orderReceived = true
				break
			}
		}
	}
	if orderReceived {
		return tempOrder
	} else {
		tempOrder.floor = -1
		return tempOrder
	}

}

func sendOrder(order ElevatorOrder) {
	if order.floor != -1 {
		fmt.Println("Type of order", order.orderType)
		fmt.Println("Floor", order.floor)
	}

}

func main() {
	init_userIO()
	for {
		newOrder := receive_order()
		sendOrder(newOrder)
	}
}
