package main

import (
	"fmt"
	"os/exec"
	"net"
	"time"
	"encoding/binary"
)

const bcast = "129.241.187.255"
const udpPort = "20015"

func main() {
	var master bool = false
	var currentNum uint64 = 0
	
	udpaddr, _ := net.ResolveUDPAddr("udp", net.JoinHostPort(bcast, udpPort))
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil { fmt.Println("Error") }

	fmt.Println("I am now Backup")
	udpmessage := make([]byte,8)
	for !(master){
		
		conn.SetReadDeadline(time.Now().Add(time.Second*2))
		
		n,_, err := conn.ReadFromUDP(udpmessage)
		
		
		if err == nil {
			currentNum = binary.BigEndian.Uint64(udpmessage[0:n])
			fmt.Println(currentNum)
		} else {
			master = true
		}
	}
	conn.Close()	
	
	fmt.Println("I am now Master")
	spawnMaster()
	conn, _ = net.DialUDP("udp", nil ,udpaddr)	
		
	for { 
		
		fmt.Println(currentNum)
		currentNum++
		binary.BigEndian.PutUint64(udpmessage, currentNum)
		_, _ = conn.Write(udpmessage)
		
		time.Sleep(time.Second)
	}

}


func spawnMaster(){ 
	 
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c" , "go run ex06.go") 
	_ = cmd.Run() 
}

