package elevator

import(
	"time"
	"driver"
	"reflect"
)

var lightArray [driver.N_BUTTONS][driver.N_FLOORS]int 

func lightThread() {
	for {
		time.Sleep(time.Millisecond * 10)
		elevatorMatrixMutex.Lock()
		setLights(getLightArray())
		elevatorMatrixMutex.Unlock()
	}
}

func setLights(lightArray [driver.N_BUTTONS][driver.N_FLOORS]int) {
	driver.Elev_set_floor_indicator(currentFloor)
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			driver.Elev_set_button_lamp(driver.Elev_button_type_t(i), j, lightArray[i][j])
		}
	}
}

func getLightArray() [driver.N_BUTTONS][driver.N_FLOORS]int { 
	var tempMatrix = make(map[string]control.ElevatorNode)
	var tempArray [driver.N_BUTTONS][driver.N_FLOORS]int
	copyMapByValue(elevatorMatrix, tempMatrix)
	for j := 0; j < driver.N_FLOORS; j++ {
		localOrders := tempMatrix[control.LocalAddress]
		tempArray[2][j] = BoolToInt(localOrders.CurrentOrders[2][j])
		for i := 0; i < driver.N_BUTTONS-1; i++ {
			for _, matrix := range tempMatrix {
				tempArray[i][j] = BoolToInt(matrix.CurrentOrders[i][j] || IntToBool(tempArray[i][j]))
			}
		}
	}
	return tempArray
}
