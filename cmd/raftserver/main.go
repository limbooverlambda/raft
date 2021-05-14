package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"

	"github.com/kitengo/raft/internal/actors"
	"github.com/kitengo/raft/internal/locator"
	"github.com/kitengo/raft/internal/models"
	"github.com/kitengo/raft/internal/server"
)

//TODO: Add the peer logic
//TODO: Create the sender stubs that will be used by both the raftclient as well as the server
//TODO: Move the models that will be used by server and client to pkg.
//TODO: Add the gob serialization and de-serialization logic
//TODO: Add the transport to receive the TCP requests and its to the appropriate channels
//TODO: Start the committer thread

func main() {
	fmt.Println("Starting raft...")
	//Initializing the service locator
	svcLocator := locator.NewServiceLocator()

	//Initializing the RaftSupervisor
	raftSupervisor := actors.NewRaftSupervisor(svcLocator)
	go func() { raftSupervisor.Start() }()

	//Initializing the raft server
	raftServer := server.NewRaftServer(svcLocator.GetRpcLocator())
	l, err := net.Listen("tcp", "127.0.0.1:4546")
	if err != nil {
		log.Printf("Error %v\n", err)
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
		go handleConnection(conn, raftServer)
	}
}

func handleConnection(conn net.Conn, raftServer server.RaftServer) {
	defer conn.Close()
	for {
		requestDecoder := gob.NewDecoder(conn)
		responseEncoder := gob.NewEncoder(conn)
		var request models.Request
		err := requestDecoder.Decode(&request)
		if err != nil {
			response := models.Response{Payload: []byte("failed..")}
			responseEncoder.Encode(&response)
			return
		}
		log.Printf("Received request %v\n", request)
		respChan, errChan := raftServer.Accept(request)
		for {
			select {
			case resp := <-respChan:
				{
					err := responseEncoder.Encode(resp)
					if err != nil {
						log.Printf("Unable to encode response %v\n", err)
					}
					return
				}
			case respErr := <-errChan:
				{
					log.Printf("Failed to process request %v\n", respErr)
					return
				}
			}
		}
	}
}
