package elevator

import(
	"control"
	"driver"
	"sync"
)

var currentDirection int
var openSendChan bool = false
var elevatorOrderMap map[string]control.ElevatorNode
var elevatorOrderMapMutex = &sync.Mutex{}


const (
	Downward = -1
	Still    = 0
	Upward   = 1
)

const (
	UpIndex       = 0
	DownIndex     = 1
	InternalIndex = 2
)

func elevatorModuleInit() {
	elevatorOrderMap = make(map[string]control.ElevatorNode)
	orderMapBeingHandled = make(map[string]control.ElevatorNode)
	for i := 0; i < driver.N_BUTTONS; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			lightArray[i][j] = 0
		}
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < driver.N_FLOORS; j++ {
			orderArray[i][j] = false
		}
	}
	driver.Elev_init()

	floor := getCurrentFloor()
	for floor == -1 {
		setElevatorDirection(driver.DIRN_DOWN)
		floor = getCurrentFloor()
	}
	setElevatorDirection(driver.DIRN_STOP)
	currentFloor = floor
	driver.Elev_set_floor_indicator(currentFloor)
	currentDirection = Still
}


func Run(sendChannel chan map[string]control.ElevatorNode, receiveChannel chan map[string]control.ElevatorNode) {
	wg := new(sync.WaitGroup)
	wg.Add(3)
	elevatorModuleInit()

	go lightThread()
	go elevatorMovementThread()
	go communicationWithControlThread(sendChannel, receiveChannel)
	wg.Wait()
}
