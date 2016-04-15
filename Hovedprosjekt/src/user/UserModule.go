package user

import (
	"driver"
	//"fmt"
	"sync"
	"time"
)

type ElevatorOrder struct {
	OrderType driver.Elev_button_type_t //Down order = 0 Up order = 1, Internal order = 2
	Floor     int                       //0 indexed(Floor 1 = 0, Floor 2 = 1 ...)
}

func userModuleInit() {
	driver.Elev_init()
}

func receiveOrder(commChannel chan ElevatorOrder) {
	var prevOrderMatrix [driver.N_FLOORS][driver.N_BUTTONS]int
	for i := 0; i < driver.N_FLOORS; i++ {
		for j := 0; j < driver.N_BUTTONS; j++ {
			prevOrderMatrix[i][j] = 0
		}
	}
	var currentOrderMatrix [driver.N_FLOORS][driver.N_BUTTONS]int
	var tempOrder ElevatorOrder
	for {
		time.Sleep(time.Millisecond * 10)
		for i := 0; i < driver.N_FLOORS; i++ {
			for j := 0; j < driver.N_BUTTONS; j++ {
				currentOrderMatrix[i][j] = driver.Elev_get_button_signal(driver.Elev_button_type_t(j), i)
			}
		}
		for i := 0; i < driver.N_FLOORS; i++ {
			for j := 0; j < driver.N_BUTTONS; j++ {
				if (currentOrderMatrix[i][j] == 1) && (prevOrderMatrix[i][j] == 0) {
					tempOrder.OrderType = driver.Elev_button_type_t(j)
					tempOrder.Floor = i
					commChannel <- tempOrder
				}
			}
		}
		prevOrderMatrix = currentOrderMatrix
	}
}

func Run(c chan ElevatorOrder) {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	userModuleInit()
	go receiveOrder(c)

	wg.Wait()
}
