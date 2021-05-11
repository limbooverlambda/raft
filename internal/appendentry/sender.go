package appendentry

import (
	"github.com/kitengo/raft/internal/models"
	client "github.com/kitengo/raft/internal/sender"
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
	MemberID     string
	RespChan     chan<- models.AppendEntryResponse
	Term         int64
	LeaderID     string
	PrevLogIndex uint64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit uint64
	MemberAddr   string
}

//Sender
//TODO: If there is a change in leadership, the go routine will be terminated. Add context for that
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
		entry := entry
		go func() {
			var err error
			defer func() {
				if err != nil {
					delayedAction(time.Second, func() {
						bufferChan <- entry
					})
				}
			}()
			log.Printf("Forwarding entry to peer %v\n", entry)
			aePayload := models.AppendEntryPayload{
				Term:         entry.Term,
				LeaderId:     entry.LeaderID,
				PrevLogIndex: entry.PrevLogIndex,
				PrevLogTerm:  entry.PrevLogTerm,
				Entries:      entry.Entries,
				LeaderCommit: entry.LeaderCommit,
			}
			resp, err := client.SendCommand(&aePayload, entry.MemberAddr)
			if err != nil {
				//Requeue the entry back into the buffer channel
				log.Printf("Unable to send AppendEntry request, retrying")
				return
			}
			var aeResp models.AppendEntryResponse
			err = aeResp.FromPayload(resp.Payload)
			if err != nil {
				log.Printf("Unable to decode AppendEntry response, retrying")
				return
			}
			entry.RespChan <- aeResp
		}()
	}
}

func delayedAction(second time.Duration, f func()) {
	ticker := time.NewTicker(second)
	defer ticker.Stop()
	<-ticker.C
	f()
}

type sender struct {
	bufferChan chan Entry
}

func (s *sender) ForwardEntry(entry Entry) {
	s.bufferChan <- entry
}
