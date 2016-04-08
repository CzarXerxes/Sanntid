package main

import (
	//"control"
	//"encoding/binary"
	//"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"
)

var elevatorTracking = make(map[string]int)
var elevatorConnections = make(map[string]net.Conn) //Dictionary with ipAddress:connectionSocket

const port = "30000"

func routerModuleInit() {
	spawnBackup()
}

func spawnBackup() {
	fmt.Println("Made a backup")
}

func assignElevatorAddress(conn net.Conn) {
	elevatorConnections[conn.RemoteAddr().String()] = conn
}

func connectNewElevatorsThread() {
	//addr, _ := net.ResolveTCPAddr("tcp", port)
	ln, _ := net.Listen("tcp", ":30000")
	for {
		var connection = *new(net.Conn)
		fmt.Println(elevatorConnections)
		connection, _ = ln.Accept()
		fmt.Println("Found elevator")
		//time.Sleep(time.Second * 1)
		assignElevatorAddress(connection)
		addNewElevatorsToTracking(connection)
	}
}

func decrementElevatorTracking(c1 chan int, cdone chan int) {
	timeStamp := time.NewTicker(time.Millisecond * 100)
	defer timeStamp.Stop()
	for _ = range timeStamp.C {
		for elevator, _ := range elevatorConnections {
			<-c1
			elevatorTracking[elevator]--
			c1 <- 1
		}
	}
}

func incrementElevatorTrackingIfAlive(c1 chan int, cdone chan int, conn *net.UDPConn) {
	buff := make([]byte, 1600)
	for {
		_, _, _ = conn.ReadFromUDP(buff)
		elevatorSlice := buff[:8]
		//elevator := int(uint64(elevatorSlice))
		elevator := string(elevatorSlice)
		<-c1
		elevatorTracking[elevator]++
		c1 <- 1
	}
}

func addNewElevatorsToTracking(conn net.Conn) {
	elevatorTracking[conn.RemoteAddr().String()] = 30
}

func elevatorIsDead(elevator string) {
	delete(elevatorTracking, elevator)
	elevatorConnections[elevator].Close()
	delete(elevatorConnections, elevator)
}

func checkIfElevatorAlive() {
	for {
		//addNewElevatorsToTracking()
		for elevator, _ := range elevatorConnections {
			if elevatorTracking[elevator] <= 0 {
				elevatorIsDead(elevator)
			}
		}
	}
}

func checkElevatorsAliveThread() {
	laddr, _ := net.ResolveUDPAddr("udp", net.JoinHostPort("", port))
	rcv, _ := net.ListenUDP("udp", laddr)
	defer rcv.Close()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	c1 := make(chan int, 1)
	cdone := make(chan int)

	go decrementElevatorTracking(c1, cdone)
	go incrementElevatorTrackingIfAlive(c1, cdone, rcv)
	c1 <- 1
	go checkIfElevatorAlive()
	/*Maybe do this
	<- cdone
	<- cdone
	*/
	wg.Wait()
}

func checkBackupAliveThread() {
	for {
		time.Sleep(10 * time.Second)
		fmt.Println("Backup is alive")
	}

}

func tellBackupAliveThread() {
	for {
		time.Sleep(10 * time.Second)
		fmt.Println("I am alive")
	}

}

func transferMatrixThread() {
	fmt.Println("Transferring matrix")
	/*
	matrixInTransit := &map[int]control.ElevatorNode{}
	for {
		for fromElevator, _ := range elevatorConnections {
			deadline := time.Now().Add(5 * time.Millisecond)
			for time.Now().Before(deadline) {
				decoderConnection := gob.NewDecoder(elevatorConnections[fromElevator])
				decoderConnection.Decode(matrixInTransit)
				for toElevator, _ := range elevatorConnections {
					if toElevator != fromElevator {
						encoderConnection := gob.NewEncoder(elevatorConnections[toElevator])
						encoderConnection.Encode(matrixInTransit)
					}
				}
			}
		}
	}
	*/
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(4)
	routerModuleInit()

	go connectNewElevatorsThread()
	go checkElevatorsAliveThread()
	go checkBackupAliveThread()
	go tellBackupAliveThread()
	go transferMatrixThread()

	wg.Wait()
}
