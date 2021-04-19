package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

//TODO: Add the transport to receive the TCP requests and it to the appropriate channels
//TODO: Start the committer thread

func main() {
	fmt.Println("Starting raft...")
	l, err := net.Listen("tcp", "127.0.0.1:4546")
	if err != nil {
		log.Printf("Error %v\n",err)
		panic("Unable to start Raft server")
	}
	defer func() {
		err := l.Close()
		if err != nil {
			log.Printf("Unable to close connection due to %v\n", err)
		}
	}()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Connection failed due to %v\n", err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		s, err:= bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Existing due to error")
			return
		}
		log.Println("received", s)
	}
}
