package user

import (
	"driver"
	//"fmt"
	"time"
)

type ElevatorOrder struct {
	OrderType driver.Elev_button_type_t //Down order = 0 Up order = 1, Internal order = 2
	Floor     int                       //0 indexed(Floor 1 = 0, Floor 2 = 1 ...)
}

func userModuleInit() {
	driver.Elev_init()
}

func receiveOrder() ElevatorOrder {
	var tempOrder ElevatorOrder
	var orderReceived bool = false
	for {
		time.Sleep(time.Millisecond * 10)
		for i := 0; i < driver.N_FLOORS; i++ {
			for j := 0; j < driver.N_BUTTONS; j++ {
				if driver.Elev_get_button_signal(driver.Elev_button_type_t(j), i) == 1 {
					tempOrder.OrderType = driver.Elev_button_type_t(j)
					tempOrder.Floor = i
					orderReceived = true
					time.Sleep(time.Millisecond * 100)
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
	userModuleInit()
	for {
		newOrder := receiveOrder()
		c <- newOrder
	}
}
