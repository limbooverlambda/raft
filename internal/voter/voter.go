package voter

import (
	raftlog "github.com/kitengo/raft/internal/log"
	raftmember "github.com/kitengo/raft/internal/member"
	raftmodels "github.com/kitengo/raft/internal/models"
	"github.com/kitengo/raft/internal/sender"
	"github.com/kitengo/raft/internal/term"
	"log"
)

type VoteStatus int

const (
	Leader VoteStatus = iota
	Follower
	Split
)

type RaftVoter interface {
	RequestVote(term int64) <-chan VoteStatus
}

func NewRaftVoter() RaftVoter {
	return &raftVoter{}
}

type raftVoter struct {
	raftMember raftmember.RaftMember
	raftLog    raftlog.RaftLog
	raftTerm   term.RaftTerm
}

func (rv *raftVoter) RequestVote(term int64) <-chan VoteStatus {
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
		defer close(voteStatusChan)
		defer close(errorChan)
		defer close(voteResponseChan)
		var voteCount int
		majorityCount := (len(members) >> 1) + 1
		for _, member := range members {
			go rv.requestVote(member, requestVotePayload, voteResponseChan, errorChan)
		}
		for {
			select {
			case vr := <-voteResponseChan:
				{
					if vr.Term > rv.raftTerm.GetTerm() {
						voteStatusChan <- Follower
						return
					}
					if vr.VoteGranted {
						voteCount++
					}
					if voteCount > majorityCount {
						voteStatusChan <- Leader
						return
					}
				}
			case err := <-errorChan:
				{
					log.Println("failed to recieve vote", err)
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
	resp, err := sender.SendCommand(&payload, member.Address)
	if err != nil {
		log.Println("unable to send vote request due to", err)
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
