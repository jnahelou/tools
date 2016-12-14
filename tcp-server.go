package main

import "net"
import "fmt"
import "strconv"
import "sync"
import "bufio"

func listenandprint(port int) {
        fmt.Println("Start socket " + strconv.Itoa(port))
        ln, _ := net.Listen("tcp", ":"+strconv.Itoa(port))
        for {
            conn, _ := ln.Accept()
            go func(c net.Conn) {
                message, _ := bufio.NewReader(c).ReadString('\n')
                fmt.Print("Message Received on " + strconv.Itoa(port) + " :", string(message))
                c.Close()
            }(conn)
        }
}
func main() {
        fmt.Println("Launching server...")
        go listenandprint(5001)
        go listenandprint(5002)
        go listenandprint(5003)
        go listenandprint(5004)
        var wg sync.WaitGroup
        wg.Add(1)
        wg.Wait()
}

