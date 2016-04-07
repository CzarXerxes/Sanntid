package main

import (
	"control"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"
)

var elevatorTracking = make([]int, 100)
var elevatorConnections = make(map[int]net.Conn) //Dictionary with assignedAddress:connectionSocket

const port = "30000"

func routerModuleInit() {
	for i := 0; i < 100; i++ {
		elevatorTracking[i] = *new(int)
	}
	spawnBackup()
}

func spawnBackup() {
	fmt.Println("Made a backup")
}

func assignElevatorAddress(conn net.Conn) {
	for i := 0; i < 100; i++ {
		fmt.Println(i)
		continueVar := false
		for elevator, _ := range elevatorConnections {
			if i == elevator {
				continueVar = true
				break
			}
		}
		if continueVar {
			continue
		}
		elevatorConnections[i] = conn
		break
	}
}

func connectNewElevatorsThread() {
	//addr, _ := net.ResolveTCPAddr("tcp", port)
	ln, _ := net.Listen("tcp", ":30000")
	for {
		fmt.Println(elevatorConnections)
		connection, _ := ln.Accept()
		fmt.Println("Found elevator")
		//time.Sleep(time.Second * 1)
		assignElevatorAddress(connection)
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
		elevator := int(binary.BigEndian.Uint64(elevatorSlice))
		<-c1
		elevatorTracking[elevator]++
		c1 <- 1
	}
}

func addNewElevatorsToTracking() {
	for elevator, _ := range elevatorConnections {
		if elevatorTracking[elevator] == *new(int) {
			elevatorTracking[elevator] = 30
		}
	}
}

func elevatorIsDead(elevator int) {
	elevatorTracking[elevator] = *new(int)
	elevatorConnections[elevator].Close()
	delete(elevatorConnections, elevator)
}

func checkIfElevatorAlive() {
	for {
		addNewElevatorsToTracking()
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
