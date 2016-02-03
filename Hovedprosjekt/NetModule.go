//This file defines a network module

package tcp

import(
	"net"
	"encoding/gob"
	"fmt"
)

type NetModule struct{
	localIP, routerIP, routerPort string
	allocatedAddress int//The address allocated by the router module when elevator has connected
	TCPSendSocket, TCPListenSocket net.TCPConn 
}


//########################### Receive packets functions #################

//This function decodes the received data from TCP as an integer and places it in 
func (n NetModule) TCPReceiveInteger(conn net.TCPConn, receivingVar int)
	dec := gob.NewDecoder(conn)
	integer := &int{}
	dec.Decode(integer)
	receivingVar = integer



//#########################################################################

func (n NetModule) TCPSendPacket(conn net.TCPConn, //datatype)
	enc := gob.NewEncoder(conn)
	enc.Encode(//)
	

func (n NetModule) TCP_Init()
	n.TCPSendSocket, err := net.Dial("tcp", net.JoinHostPort(n.routerIP, n.routerPort))
	if err != nil{
		fmt.PrintLn("Failed to open TCP sending socket\n")
		return err
	}

	n.TCPListenSocket, err = net.Listen("tcp", ":" + n.routerPort)
	if err != nil{
		fmt.PrintLn("Failed to open TCP listening socket\n")
		return err
	}
	
	conn, err := n.TCPListenSocket.Accept()
	if err != nil{
		fmt.PrintLn("Failed to accept listen request\n")
		return err
	}
	
	go TCPReceiveInteger(conn, n.allocatedAddress)//

func (n NetModule) TCP_Destroy()
	n.TCPSendSocket.Close()
	n.TCPListenSocket.Close()
	
	
