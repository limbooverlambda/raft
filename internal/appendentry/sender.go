package appendentry

import (
	"context"
	raftlog "github.com/kitengo/raft/internal/log"
	"github.com/kitengo/raft/internal/models"
	raftsender "github.com/kitengo/raft/internal/sender"
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
	MemberPort   string
}

//Sender
// Used to send appendEntry RPCs to the peers
type Sender interface {
	ForwardEntry(entry Entry)
}

func NewSender(ctxt context.Context, raftSender raftsender.RequestSender, raftLog raftlog.RaftLog) Sender {
	bufferChan := make(chan Entry, 100)
	go senderSiphon(ctxt, raftSender, raftLog, bufferChan)
	return &sender{
		bufferChan: bufferChan,
	}
}

func senderSiphon(ctxt context.Context, raftSender raftsender.RequestSender, raftLog raftlog.RaftLog, bufferChan chan Entry) {
	cctxt, cancel := context.WithCancel(ctxt)
	for {
		select {
		case <-ctxt.Done():
			log.Println("Terminating the child routines")
			cancel()
			return
		case entry := <-bufferChan:
			go sendAppendEntryReq(cctxt, bufferChan, entry, raftSender, raftLog)
		}
	}
}

func sendAppendEntryReq(ctxt context.Context, bufferChan chan Entry, entry Entry, raftSender raftsender.RequestSender, raftLog raftlog.RaftLog) {
	select {
	case <-ctxt.Done():
		log.Println("Cancelling the child sender and closing the response channel")
		close(entry.RespChan)
		return
	default:
		var err error
		defer func() {
			if err != nil {
				delayedAction(time.Second, func() {
					log.Println("failed send, retry")
					select {
					case <-ctxt.Done():
						{
							log.Println("context cancelled, no bufferring needed")
							close(entry.RespChan)
						}
					default:
						log.Println("buffering the failure")
						bufferChan <- entry
					}
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
		resp, err := raftSender.SendCommand(&aePayload, entry.MemberAddr, entry.MemberPort)
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
		if !aeResp.Success {
			//Append failed, decrement the index and retry
			nextIndex := entry.PrevLogIndex - 1
			nextEntry, err := raftLog.GetLogEntryAtIndex(nextIndex)
			if err != nil {
				log.Printf("Unable to AppendEntry after decrementing index, retrying")
				return
			}
			modifiedEntry := Entry{
				MemberID:     entry.MemberID,
				RespChan:     entry.RespChan,
				Term:         entry.Term,
				LeaderID:     entry.LeaderID,
				PrevLogIndex: nextEntry.Index,
				PrevLogTerm:  int64(nextEntry.Term),
				Entries:      nextEntry.Payload,
				LeaderCommit: entry.LeaderCommit,
				MemberAddr:   entry.MemberAddr,
			}
			log.Println("Sending modified entry")
			sendToChan(ctxt, func() {
				bufferChan <- modifiedEntry
			})
		}
		log.Println("Sending response over channel")
		sendToChan(ctxt, func() {
			entry.RespChan <- aeResp
		})
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

func sendToChan(ctx context.Context,
	sendFn func(),
) {
	select {
	case <-ctx.Done():
		return
	default:
		sendFn()
		return
	}
}
