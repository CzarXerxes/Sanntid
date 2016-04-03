package main

import (
    "fmt"
    "net"
    "time"
    "runtime"
    //"os"
    //"strconv"
    "sync"
    //"strings"
)

const localIP = "129.241.187.148"
const bcast = "129.241.187.255"//Server IP
const udpListen = "30000"//Local listen port
const udpSend = "20015" //Local send port

//Error function

func CheckError(err error){
	if err != nil{
		fmt.Println("Error: ", err)
	}
}


//UDP Connector

func ConUDP(listenPort string, sendPort string){
	///////////////////////////////////////////
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	var wg sync.WaitGroup
	wg.Add(2)
	//////////////////////////////////////////
		
	
	
	go listenThread(listenPort)
	go sendThread(sendPort)
	go listenThread(sendPort)
	
	wg.Wait()
	
	

}

func listenThread(port string){
	laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("",port))//laddr = "listen address"
	CheckError(err)	
	conn, err := net.ListenUDP("udp", laddr)//rcv = "receive object"
	CheckError(err)

	fmt.Println("Listening started...")

	buff := make( []byte, 1600)
	for i:=0; i<10; i++{
		_,_,err := conn.ReadFromUDP(buff)
		CheckError(err)
		str := string(buff)
		fmt.Println(str)
		time.Sleep(time.Millisecond*100)
	}
}
		

func sendThread(port string){
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(bcast, port))//raddr = "receive address"
	CheckError(err)
	
	conn, err := net.DialUDP("udp", nil, raddr)//send = "send object"
	CheckError(err)

	fmt.Println("Sending started...")
	for i:=0;i<10;i++{
		_, err := fmt.Fprintf(conn, "Message")
		CheckError(err)
		time.Sleep(time.Millisecond*100)
	}
}

func main(){
	ConUDP(udpListen, udpSend)
}
