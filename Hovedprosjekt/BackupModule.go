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

const IP = "129.241.187.153" //Start router on this IP

var elevatorConnections = make(map[string]net.Conn)

var routerIsDead bool

//var routerTCPConnection net.Conn
var routerAliveConnection net.Conn
var routerCommConnection net.Conn
var routerDecoder *gob.Decoder
var routerIPAddress string

var routerPort = "30000"
var elevatorPort = "28000"

func getRouterTCPConnection() {
	var err error
	fmt.Println("Connecting to router")
	time.Sleep(time.Second * 1)
	getRouterIP()
	routerAliveConnection, err = net.Dial("tcp", net.JoinHostPort(routerIPAddress, routerPort))
	if err != nil {
		fmt.Println("There has been an error. I am not connected to router")
	}
	time.Sleep(time.Millisecond * 200)
	routerCommConnection, _ = net.Dial("tcp", net.JoinHostPort(routerIPAddress, routerPort))
	routerDecoder = gob.NewDecoder(routerCommConnection)
	fmt.Println("Connected to router")
}

func backupInit() {
	fmt.Println("Hello. I am backup")
	getRouterTCPConnection()
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
func getRouterIP() {
	routerIPAddress = IP
}

func returnLocalIP() string { //Implement this function to get local IP
	return IP
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
			commitSuicide()
			routerIsDead = false
		}
	}
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(3)
	backupInit()
	time.Sleep(time.Second * 1)

	go tellRouterStillAliveThread()
	go checkIfRouterStillAliveThread()
	go spawnNewRouterModule()

	wg.Wait()
}
