package main

import (
    "bytes"
    "encoding/gob"
    "fmt"
    "log"
)

const server = "129.241.187.23"
const port = "33546"

type P struct {
    X, Y, Z int
    Name    string
}

func main(){
	conn, _ := net.Dial("tcp", net.JoinHostPort(server, port))
	var network bytes.Buffer	
	
	enc := gob.NewEncoder(conn)
	enc.Encode(P{3,4,5, "ASdjasdj"})	
	
	conn.Close()
}
