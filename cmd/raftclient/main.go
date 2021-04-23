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
		//Send the client request payload
		clientCommand := &raftmodels.ClientCommandPayload{
			ClientCommand: []byte(payload),
		}
		sendCommand(clientCommand,ip)
	},
}

var appendEntryCmd = &cobra.Command{
	Use: "aecmd",
	Short: "Send append entry requests to the raft server",
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")
		entries, _ := cmd.Flags().GetString("entries")
		leaderID, _ := cmd.Flags().GetString("leaderid")
		prevLogIndex, _ := cmd.Flags().GetInt64("prevlogindex")
		prevLogTerm, _ := cmd.Flags().GetInt64("prevlogterm")
		leaderCommit, _ := cmd.Flags().GetInt64("leadercommit")
		term, _ := cmd.Flags().GetInt64("term")
		fmt.Println("Sending ae command")
		aeCommand := &raftmodels.AppendEntryPayload{
			Term:         term,
			LeaderId:     leaderID,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      []byte(entries),
			LeaderCommit: leaderCommit,
		}
		sendCommand(aeCommand, ip)
	},
}

var requestVoteCmd = &cobra.Command{
	Use: "votecmd",
	Short: "Send request Vote to the raft server",
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")
		candidateID, _ := cmd.Flags().GetString("candidateid")
		lastLogIndex, _ := cmd.Flags().GetInt64("lastlogindex")
		lastLogTerm, _ := cmd.Flags().GetInt64("lastlogterm")
		term, _ := cmd.Flags().GetInt64("term")
		fmt.Println("Sending vote command")
		command := &raftmodels.RequestVotePayload{
			Term:         term,
			CandidateId:  candidateID,
			LastLogIndex: lastLogIndex,
			LastLogTerm:  lastLogTerm,
		}
		sendCommand(command, ip)
	},
}


func main() {
	clientCmd.Flags().StringP("ip","","127.0.0.1:4546", "server to send client requests to")
	clientCmd.Flags().StringP("payload","", "","string payload to send to the raft server")

	appendEntryCmd.Flags().StringP("ip","","127.0.0.1:4546", "server to send client requests to")
	appendEntryCmd.Flags().StringP("entries","", "","string payload to append to the raft server")
	appendEntryCmd.Flags().StringP("leaderid","", "","leader id for the append entry")
	appendEntryCmd.Flags().Int64P("prevlogindex","", 0,"prev log index for the append entry")
	appendEntryCmd.Flags().Int64P("prevlogterm","", 0,"prev log term for the append entry")
	appendEntryCmd.Flags().Int64P("leadercommit","", 0,"leader commit for the append entry")
	appendEntryCmd.Flags().Int64P("term","", 0,"term for the append entry")

	requestVoteCmd.Flags().StringP("ip","","127.0.0.1:4546", "server to send client requests to")
	requestVoteCmd.Flags().StringP("candidateid","", "","candidate id for the request vote")
	requestVoteCmd.Flags().Int64P("lastlogindex","", 0,"last log index for the request vote")
	requestVoteCmd.Flags().Int64P("lastlogterm","", 0,"last log term for the request vote")
	requestVoteCmd.Flags().Int64P("term","", 0,"term for the request vote")

	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(appendEntryCmd)
	rootCmd.AddCommand(requestVoteCmd)

	rootCmd.Execute()
}

func sendCommand(requestConv raftmodels.RequestConverter, ipAddress string) {
	//TODO: Send the client request, appendEntry and voteRequest to the server
	//TODO: On the server side ensure that the payload is decoded
	//TODO: Wire up the RaftServer with the TCP server
	//TODO: Have the mocks return happy path payloads for the local setup
	conn, err := net.Dial("tcp", ipAddress)
	if err != nil {
		log.Panic("Unable to dial due to", err)
	}
	defer conn.Close()
	req, err := requestConv.ToRequest()
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


