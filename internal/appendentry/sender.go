package appendentry

import (
	"log"
	"time"
)

//Entry
// term leader’s term
// leaderId so follower can redirect clients
// prevLogIndex index of log entry immediately preceding
// new ones
// prevLogTerm term of prevLogIndex entry
// entries[] log entries to store (empty for heartbeat;
//may send more than one for efficiency)
//leaderCommit leader’s commitIndex
type Entry struct {
	MemberID string
	RespChan chan <- Response
}

//Response
//term currentTerm, for leader to update itself
//success true if follower contained entry matching
//prevLogIndex and prevLogTerm
type Response struct {
	Term    int64
	Success bool
}
//Sender
//TODO: Create a struct for the mailbox, this will be the retry buffer
//TODO: Create the Add the AppendEntry payload to the send mailbox. Create the sender interface
//TODO: Spawn a go-routine that siphons messages from the send mailbox
//TODO: Create the entry struct to be used by the sender. If the send is successful, send the success response into the entries success channel
//TODO: If the send is a failure, mitigate by requeuing the payload back to the retry buffer
//If there is a change in leadership, the go routine will be terminated
type Sender interface {
	ForwardEntry(entry Entry)
}

func NewSender() Sender {
	bufferChan := make(chan Entry, 100)
	go senderSiphon(bufferChan)
	return &sender{
		bufferChan: bufferChan,
	}
}

func senderSiphon(bufferChan chan Entry) {
	for entry := range bufferChan {
		go func() {
			log.Printf("Forwarding entry to peer %v\n", entry)
			time.Sleep(300 * time.Millisecond)
			entry.RespChan <- Response{
				Term:    0,
				Success: true,
			}
		}()
	}
}

type sender struct{
	bufferChan chan Entry
}

func (s *sender) ForwardEntry(entry Entry) {
	s.bufferChan <- entry
}

