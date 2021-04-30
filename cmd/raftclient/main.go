package main

import (
	"fmt"
	raftmodels "github.com/kitengo/raft/internal/models"
	raftsender "github.com/kitengo/raft/internal/sender"
	"github.com/spf13/cobra"
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
		raftsender.SendCommand(clientCommand,ip)
	},
}

var appendEntryCmd = &cobra.Command{
	Use: "aecmd",
	Short: "Send append entry requests to the raft server",
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")
		entries, _ := cmd.Flags().GetString("entries")
		leaderID, _ := cmd.Flags().GetString("leaderid")
		prevLogIndex, _ := cmd.Flags().GetUint64("prevlogindex")
		prevLogTerm, _ := cmd.Flags().GetInt64("prevlogterm")
		leaderCommit, _ := cmd.Flags().GetUint64("leadercommit")
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
		raftsender.SendCommand(aeCommand, ip)
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
		raftsender.SendCommand(command, ip)
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



