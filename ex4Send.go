package main

import (
    "net"
    "encoding/gob"
)

const server = "129.241.187.144"
const port = "30000"

type P struct {
    X, Y, Z int
    Name    string
}

func main(){
	conn, _ := net.Dial("tcp", net.JoinHostPort(server, port))
	
	enc := gob.NewEncoder(conn)
	enc.Encode(P{3,4,5, "ASdjasdj"})	
	
	conn.Close()
}
