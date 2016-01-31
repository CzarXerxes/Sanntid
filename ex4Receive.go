package main

import (
    "net"
    "encoding/gob"
    "fmt"
)

type P struct {
    X, Y, Z int
    Name    string
}

func handleConnection(conn net.Conn){
	dec := gob.NewDecoder(conn)
	p := &P{}
	dec.Decode(p)
	fmt.Printf("%d, %d, %d, %q", p.X, p.Y, p.Z, p.Name)

func main(){
	ln, _ := net.Listen("tcp", ":30000")
	
	for{
		conn, _ := ln.Accept()
		go handleConnection(conn)
	}
}
