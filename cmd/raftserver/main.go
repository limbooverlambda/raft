package main

import (
	"encoding/gob"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net"

	"github.com/kitengo/raft/internal/actors"
	"github.com/kitengo/raft/internal/locator"
	"github.com/kitengo/raft/internal/models"
	"github.com/kitengo/raft/internal/server"
)

//TODO: Start the committer thread
var rootCmd = &cobra.Command{
	Use:   "raftserver",
	Short: "Raft server binary",
}

var serverStartupCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the raft server",
	Run: func(cmd *cobra.Command, args []string) {
		peers, _ := cmd.Flags().GetStringSlice("peers")
		port, _ := cmd.Flags().GetInt16("port")
		name, _ := cmd.Flags().GetString("name")
		log.Println("Peers", peers)
		log.Println("port", port)
		log.Println("name", name)
	},
}

func main() {
	var peers []string
	var defaultPeers []string
	serverStartupCmd.Flags().StringSliceVarP(&peers, "peers", "p",
		defaultPeers, "--peers <peer1> --peers <peer2> ...")

	var port int16
	defaultPort := int16(8326)
	serverStartupCmd.Flags().Int16VarP(&port, "port", "r",
		defaultPort, "--port <port number of the server>")

	var name string
	//Randomize the default name
	defaultName := "node"
	serverStartupCmd.Flags().StringVarP(&name, "name", "n",
		defaultName, "--name <name of the server>")

	rootCmd.AddCommand(serverStartupCmd)
	rootCmd.Execute()
}

func startServer() {
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
