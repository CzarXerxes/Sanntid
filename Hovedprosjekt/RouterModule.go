package main

import (
	//"control"
	//"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"
	//"bufio"
	"os/exec"
)

var elevatorTracking = make(map[string]bool)
var elevatorConnections = make(map[string]net.Conn) //Dictionary with ipAddress:connectionSocket

var backupIPAddress = "129.241.187.152"
var backupConnection net.Conn
var backupIsDead bool

const port = ":30000"

func routerModuleInit() {
	spawnBackup()
}
/*
func getTCPBackupConnection(){
	fmt.Println("Connecting to backup")
	backupConnection, _ = net.Dial("tcp", net.JoinHostPort(backupIPAddress, port))
	fmt.Println("Connected to backup")
}
*/

func spawnBackup() {
	fmt.Println("Making a backup")
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c" , "go run BackupModule.go")
	_ = cmd.Run()
	//Try this
	ln, _ := net.Listen("tcp", port)
	for{
		backupConnection,_ = ln.Accept()
		fmt.Println("Connected to backup")	
	}
	time.Sleep(time.Second * 5)
	//
	/*
	time.Sleep(time.Second * 5)
	go getTCPBackupConnection()
	time.Sleep(time.Millisecond * 500)
	*/
}

func sendElevatorMapToBackup(){
	enc := gob.NewEncoder(backupConnection)
	enc.Encode(elevatorConnections)
}

//Fix this bug: Adding new conn to map overwrites previous conn
func assignElevatorAddress(conn net.Conn) {
	elevatorConnections[conn.RemoteAddr().String()] = conn
	sendElevatorMapToBackup()
}

func addNewElevatorsToTracking(conn net.Conn) {
	elevatorTracking[conn.RemoteAddr().String()] = true
}


func connectNewElevatorsThread() {
	//addr, _ := net.ResolveTCPAddr("tcp", port)
	ln, _ := net.Listen("tcp", port)
	for{

		//ln, _ := net.Listen("tcp", ":30000")
		//var connection = *new(net.Conn)
		fmt.Println(elevatorConnections)
		fmt.Println(elevatorTracking)
		connection, _ := ln.Accept()
		fmt.Println("Found elevator")
		//time.Sleep(time.Second * 1)
		assignElevatorAddress(connection)
		addNewElevatorsToTracking(connection)
	}
}

/*
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
*/


//Other errors than cut network connection could kill elevator
func elevatorStillConnected(elevatorIP string) bool{
	socket := elevatorConnections[elevatorIP]
	buf := make([]byte, 1024)
	_, err := socket.Read(buf)
	if err != nil {
		return false
	}
	//fmt.Printf("Message received :: %s\n", string(buf[:n]))
	return true
}

func checkElevatorStillConnectedThread(){
	for{
		for elevator, _ := range elevatorConnections{
			if !elevatorStillConnected(elevator){
				elevatorIsDead(elevator)
			}
		}
	}
}
	


func elevatorIsDead(elevator string) {
	delete(elevatorTracking, elevator)
	elevatorConnections[elevator].Close()
	delete(elevatorConnections, elevator)
	fmt.Println(elevatorConnections)
}

/*
func checkIfElevatorAlive() {
	for {
		for elevator, _ := range elevatorTracking {
			if elevatorTracking[elevator] == false {
				elevatorIsDead(elevator)
			}
		}
	}
}
*/

func checkBackupAliveThread() {
	for {
		buf := make([]byte, 1024)
		_, err := backupConnection.Read(buf)
		if err != nil {
			backupIsDead =  true
		}
		backupIsDead =  false
	}
}

func tellBackupAliveThread() {
	for{
		fmt.Println("Telling backup alive")	
		time.Sleep(time.Millisecond * 100)
		//fmt.Println("Sending im alive")
		text := "Router is still alive"
		fmt.Fprintf(backupConnection, text)	
	}
}

func spawnNewBackupThread(){
	for{
		if backupIsDead{
			backupIPAddress = "129.241.187.152"
			spawnBackup()
		}
	}
}

func transferMatrixThread() {
	/*
	fmt.Println("Transferring matrix")
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
	go checkElevatorStillConnectedThread()
	go checkBackupAliveThread()
	go tellBackupAliveThread()
	go transferMatrixThread()

	wg.Wait()
}
