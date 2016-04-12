package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
)

var elevatorConnections = make(map[string]net.Conn)

var routerIsDead bool

//var routerTCPConnection net.Conn
var routerAliveConnection net.Conn
var routerCommConnection net.Conn
var routerDecoder *gob.Decoder
var routerIPAddress = "129.241.187.153"

var newRouterIP = "129.241.187.153"
var routerPort = "30000"
var elevatorPort = "28000"

func getRouterTCPConnection() {
	fmt.Println("Connecting to router")
	time.Sleep(time.Second * 1)
	routerAliveConnection, _ = net.Dial("tcp", net.JoinHostPort(routerIPAddress, routerPort))
	time.Sleep(time.Millisecond * 20)
	routerCommConnection, _ = net.Dial("tcp", net.JoinHostPort(routerIPAddress, routerPort))
	routerDecoder = gob.NewDecoder(routerCommConnection)
	fmt.Println("Connected to router")
}

func backupInit() {
	fmt.Println("Hello. I am backup")
	getRouterTCPConnection()
}

//Implement this to receive elevator list
func receiveElevatorList() {
	var decodedMap map[string]net.Conn
	for {
		time.Sleep(time.Millisecond * 100)
		routerDecoder.Decode(&decodedMap)
		elevatorConnections = decodedMap
	}
}

func tellRouterStillAliveThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		text := "Backup is still alive"
		fmt.Fprintf(routerAliveConnection, text)
	}
}

func checkIfRouterStillAliveThread() {
	for {
		time.Sleep(time.Millisecond * 100)
		buf := make([]byte, 1024)
		_, err := routerAliveConnection.Read(buf)
		if err != nil {
			routerIsDead = true
			fmt.Println("Router is dead")
		}
		routerIsDead = false
	}
}

func sendNewRouterAddressToElevators() {
	for i := 0; i < 10; i++ {
		for elevatorIPAddress, _ := range elevatorConnections {
			raddr, _ := net.ResolveUDPAddr("udp", net.JoinHostPort(elevatorIPAddress, elevatorPort))
			conn, _ := net.DialUDP("udp", nil, raddr)
			_, _ = fmt.Fprintf(conn, newRouterIP)
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func openNewRouter() {
	fmt.Println("Opening new router")
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run RouterModule.go")
	_ = cmd.Run()
}

func commitSuicide() {
	fmt.Println("Commiting suicide")
	backupPid := os.Getpid()
	backupProcess, _ := os.FindProcess(backupPid)
	backupProcess.Kill()
}

func spawnNewRouterModule() {
	for {
		if routerIsDead {
			fmt.Println("Router is dead")
			openNewRouter()
			sendNewRouterAddressToElevators()
			commitSuicide()
			routerIsDead = false
		}
	}
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(4)
	backupInit()
	time.Sleep(time.Second * 1)

	go receiveElevatorList()
	go tellRouterStillAliveThread()
	go checkIfRouterStillAliveThread()
	go spawnNewRouterModule()

	wg.Wait()
}
