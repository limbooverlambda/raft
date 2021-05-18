package main

import (
	"encoding/gob"
	"fmt"
	"github.com/kitengo/raft/internal/rconfig"
	"github.com/spf13/cobra"
	"log"
	"net"
	"strings"

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
		port, _ := cmd.Flags().GetString("port")
		name, _ := cmd.Flags().GetString("name")

		config := rconfig.Config{
			ServerID:     name,
			ServerPort:   port,
			MemberConfig: toMemberConfig(peers),
		}
		startServer(config)
	},
}

func toMemberConfig(peers []string) []rconfig.MemberConfig {
	var members []rconfig.MemberConfig
	for _, peer := range peers {
		tokens := strings.Split(peer, ":")
		mConfig := rconfig.MemberConfig{
			ID:   tokens[0],
			IP:   tokens[1],
			Port: tokens[2],
		}
		members = append(members, mConfig)
	}
	return members
}

func main() {
	var peers []string
	var defaultPeers []string
	serverStartupCmd.Flags().StringSliceVarP(&peers, "peers", "p",
		defaultPeers, "--peers <name:ip:port> --peers <name:ip:port> ...")

	var port string
	defaultPort := "4546"
	serverStartupCmd.Flags().StringVarP(&port, "port", "r",
		defaultPort, "--port <port number of the server>")

	var name string
	//Randomize the default name
	defaultName := "node"
	serverStartupCmd.Flags().StringVarP(&name, "name", "n",
		defaultName, "--name <name of the server>")

	rootCmd.AddCommand(serverStartupCmd)
	rootCmd.Execute()
}

func startServer(config rconfig.Config) {
	fmt.Println("Starting raft...")
	//Initializing the service locator
	svcLocator := locator.NewServiceLocator(config)

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
