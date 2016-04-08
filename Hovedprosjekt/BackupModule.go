package main

import(
	"fmt"
	"encoding/gob"
	"net"
	"time"
	"os/exec"
	"sync"
)

var elevatorConnections = make(map[string]net.Conn)

var routerIsDead bool
var routerTCPConnection net.Conn
var routerIPAddress = "129.241.187.152"

var newRouterIP = "129.241.187.152"
var port = "30000"

func getRouterTCPConnection(){
	fmt.Println("Connecting to router")
	routerTCPConnection, _ = net.Dial("tcp", net.JoinHostPort(routerIPAddress, port))
	fmt.Println("Connected to router")
}

func backupInit(){
	fmt.Println("Hello. I am backup")
	//Try this
	go getRouterTCPConnection()
	//time.Sleep(time.Millisecond * 500)
	/*
	ln, _ := net.Listen("tcp", port)
	routerTCPConnection,_ = ln.Accept()
	fmt.Println("Connected to router")
	*/
}


func receiveElevatorList(){
	for{
		dec := gob.NewDecoder(routerTCPConnection)
		var list = make(map[string]net.Conn)
		dec.Decode(list)
		elevatorConnections = list	
	}
}

func tellRouterStillAliveThread() {
	for{		
		time.Sleep(time.Millisecond * 100)
		//fmt.Println("Sending im alive")
		text := "Backup is still alive"
		fmt.Fprintf(routerTCPConnection, text)	
	}
}

func checkIfRouterStillAliveThread() {
	buf := make([]byte, 1024)
	_, err := routerTCPConnection.Read(buf)
	if err != nil{
		routerIsDead = true
	}
	routerIsDead = false
}


func sendNewRouterAddressToElevators(){
	for i:=0; i< 10; i++{
		for elevatorIPAddress, _ := range elevatorConnections{
			raddr, _ := net.ResolveUDPAddr("udp", net.JoinHostPort(elevatorIPAddress, port))
			conn, _ := net.DialUDP("udp", nil, raddr)
			_,_ = fmt.Fprintf(conn, newRouterIP)
			time.Sleep(time.Millisecond*100)
		}
	}
}

func openNewRouter(){
	fmt.Println("Opening new router")
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c" , "go run RouterModule.go") 
	_ = cmd.Run()
}

func commitSuicide(){
	fmt.Println("Commiting suicide")
}


func spawnNewRouterModule(){
	for{
		if routerIsDead{
			openNewRouter()
			sendNewRouterAddressToElevators()
			commitSuicide() 
		}
	}
}

func main(){
	wg := new(sync.WaitGroup)
	wg.Add(4)
	backupInit()
	time.Sleep(time.Second*5)
	go receiveElevatorList()
	go tellRouterStillAliveThread()
	go checkIfRouterStillAliveThread()
	go spawnNewRouterModule()

	wg.Wait()
}
