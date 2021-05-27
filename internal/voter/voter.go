package voter

import (
	raftlog "github.com/kitengo/raft/internal/log"
	raftmember "github.com/kitengo/raft/internal/member"
	raftmodels "github.com/kitengo/raft/internal/models"
	"github.com/kitengo/raft/internal/sender"
	raftterm "github.com/kitengo/raft/internal/term"
	"log"
)

//go:generate stringer -type=VoteStatus
type VoteStatus int

const (
	Leader VoteStatus = iota
	Follower
	Split
)

type RaftVoter interface {
	RequestVote(term int64) <-chan VoteStatus
}

func NewRaftVoter(raftMember raftmember.RaftMember,
	raftLog raftlog.RaftLog,
	raftTerm raftterm.RaftTerm) RaftVoter {
	return &raftVoter{
		raftMember: raftMember,
		raftLog:    raftLog,
		raftTerm:   raftTerm,
	}
}

type raftVoter struct {
	raftMember raftmember.RaftMember
	raftLog    raftlog.RaftLog
	raftTerm   raftterm.RaftTerm
}

func (rv *raftVoter) RequestVote(term int64) <-chan VoteStatus {
	log.Println("Requesting vote for term", term)
	voteStatusChan := make(chan VoteStatus, 1)
	candidateID := rv.raftMember.Self().ID
	lastLogEntryMeta := rv.raftLog.GetCurrentLogEntry()
	//Create the RequestVote payload
	requestVotePayload := raftmodels.RequestVotePayload{
		Term:         term,
		CandidateId:  candidateID,
		LastLogIndex: int64(lastLogEntryMeta.Index),
		LastLogTerm:  int64(lastLogEntryMeta.Term),
	}
	members, err := rv.raftMember.List()
	if err != nil {
		log.Println("Unable to list the peers, closing the vote channel")
		close(voteStatusChan)
	}

	voteResponseChan := make(chan raftmodels.RequestVoteResponse)
	errorChan := make(chan error)
	go func() {
		log.Println("Spawning go routine to request vote for term", term)
		defer func() {
			log.Println("Closing all the channels for request vote")
			close(voteStatusChan)
			close(errorChan)
			close(voteResponseChan)
		}()

		//Vote for itself
		voteCount := 1
		rv.raftMember.SetVotedFor(candidateID)

		majorityCount := (len(members) >> 1) + 1
		for _, member := range members {
			go rv.requestVote(member, requestVotePayload, voteResponseChan, errorChan)
		}
		for {
			select {
			case vr := <-voteResponseChan:
				{
					if vr.Term > rv.raftTerm.GetTerm() {
						log.Printf("Incoming term %d is greater than current term %d. Be a follower\n", vr.Term, rv.raftTerm.GetTerm())
						voteStatusChan <- Follower
						return
					}
					if vr.VoteGranted {
						voteCount++
					}
					if voteCount > majorityCount {
						log.Printf("Got the voteCount %d which is greater than the majority count %d. Be a leader\n", voteCount, majorityCount)
						rv.raftMember.SetSelfToLeader()
						voteStatusChan <- Leader
						return
					}
				}
			case err := <-errorChan:
				{
					log.Println("failed to receive vote", err)
				}
			}
		}
	}()
	return voteStatusChan
}

func (rv *raftVoter) requestVote(member raftmember.Entry,
	payload raftmodels.RequestVotePayload,
	responseChan chan raftmodels.RequestVoteResponse,
	errChan chan error) {
	log.Println("Spawned go routine to request vote from member", member.ID)
	defer func() {
		log.Println("Exiting the request vote go routine")
	}()
	resp, err := sender.SendCommand(&payload, member.Address, member.Port)
	if err != nil {
		log.Println("unable to send vote request due to", err)
		errChan <- err
		return
	}
	var vrResp raftmodels.RequestVoteResponse
	err = vrResp.FromPayload(resp.Payload)
	if err != nil {
		log.Println("unable to decode VoteRequest response", err)
		errChan <- err
		return
	}
	responseChan <- vrResp
	return
}
