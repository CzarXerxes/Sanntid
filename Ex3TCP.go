package main

import (
    "fmt"
    "net"
    "runtime"
    "bufio"
    "os"
    "strconv"
    "sync"
    "time"
)



const server = "129.241.187.23"
const tcpPortDelim = "33546"
const tcpPortListening = "30000"


// #### 0-DELIMITED TCP READER

func readAndPrintDelim(reader *bufio.Reader) {
    str, _ := reader.ReadString('\000')
    fmt.Println("received: " + str)
}



// #### TCP CONNECTOR

func conTCP(port string, callback func(*bufio.Reader)) {
    conn, err := net.Dial("tcp", net.JoinHostPort(server, port))
    if err != nil {
        fmt.Fprintln(os.Stderr, "connection error on " + server + ":" + port)
        fmt.Fprintln(os.Stderr, "Connection join error: " + err.Error())
        return
    }
    defer conn.Close()
    
    // chat withe server :-}
    wg.Add(1)
    chatWithServer(conn, callback)
    
    // initialize socket for listening
    addr, err := net.ResolveTCPAddr("tcp", ":"+tcpPortListening)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Address resolution error: " + err.Error())
    }
    ln, err := net.ListenTCP("tcp", addr)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Connection init error: " + err.Error())
    }
    ln.SetDeadline(time.Now().Add(2*time.Second))
    defer ln.Close()
    // ask for 2 reverse connection
    fmt.Fprintf(conn, "Connect to: 129.241.187.148:" + tcpPortListening + "\000")
    fmt.Fprintf(conn, "Connect to: 129.241.187.148:" + tcpPortListening + "\000")
    // handle incoming connections
    deadline := time.Now().Add(2*time.Second)
    for time.Now().Before(deadline) {
        conn2, err := ln.Accept()
        if err != nil {
            fmt.Fprintln(os.Stderr, "Socket acceptance error: " + err.Error())
            continue
        }
        defer conn2.Close()
        wg.Add(1)
        go chatWithServer(conn2, callback)
    }
    wg.Wait();
    
}
	
// #### server CHATTER

func chatWithServer(conn net.Conn, callback func(*bufio.Reader)) {
    reader := bufio.NewReaderSize(conn, 1024)
    // read welcome message
    callback(reader)
    // chat for a while
    for i := 0; i < 5; i++ {
        fmt.Fprintf(conn, "Hello world: " + strconv.Itoa(i) + "\000")
        callback(reader)
    }
    wg.Done()
}

// #### MAIN
var wg sync.WaitGroup

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    fmt.Println("=== 0-delimited TCP communication ===");
    conTCP(tcpPortDelim, readAndPrintDelim)
}
