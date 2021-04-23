package main

import (
	"encoding/gob"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net"

	raftmodels "github.com/kitengo/raft/internal/models"
)

var rootCmd = &cobra.Command{
	Use:   "raftclient",
	Short: "An executable to execise the raft APIs",
}


var clientCmd = &cobra.Command{
	Use: "clientcmd",
	Short: "Send client requests to the raft server",
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")
		payload, _ := cmd.Flags().GetString("payload")
		fmt.Printf("Sending client command %v to %v\n", payload, ip)
		clientCommand(ip, payload)
	},
}


func main() {
	clientCmd.Flags().StringP("ip","","127.0.0.1:4546", "server to send client requests to")
	clientCmd.Flags().StringP("payload","", "","string payload to send to the raft server")
	rootCmd.AddCommand(clientCmd)
	rootCmd.Execute()
}

func clientCommand(ipAddress ,payload string) {
	//TODO: Send the client request, appendEntry and voteRequest to the server
	//TODO: On the server side ensure that the payload is decoded
	//TODO: Wire up the RaftServer with the TCP server
	//TODO: Have the mocks return happy path payloads for the local setup
	conn, err := net.Dial("tcp", ipAddress)
	if err != nil {
		log.Panic("Unable to dial due to", err)
	}
	defer conn.Close()
	//Send the client request payload
	clientCommand := raftmodels.ClientCommandPayload{
		ClientCommand: []byte(payload),
	}
	req, err := clientCommand.ToRequest()
	if err != nil {
		log.Panic("Unable to serialize client comand request", err)
	}
	connEncoder := gob.NewEncoder(conn)
	err = connEncoder.Encode(req)
	if err != nil {
		log.Panic("Unable to send client command request", err)
	}
	recvChan := make(chan struct{})
	go func() {
		defer close(recvChan)
		connDecoder := gob.NewDecoder(conn)
		response := &raftmodels.Response{}
		err := connDecoder.Decode(response)
		if err != nil {
			log.Printf("Unable to decode response %v\n", err)
			return
		}
		log.Printf("Got response %v\n", string(response.Payload))
	}()
	<-recvChan
	log.Println("Sent the client command request")
}
