package voter

import (
	"fmt"
	raftlog "github.com/kitengo/raft/internal/log"
	raftmember "github.com/kitengo/raft/internal/member"
	raftmodels "github.com/kitengo/raft/internal/models"
	"github.com/kitengo/raft/internal/sender"
	raftterm "github.com/kitengo/raft/internal/term"
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
		defer close(voteStatusChan)
		defer close(errorChan)
		defer close(voteResponseChan)
		//Vote for itself
		voteCount := 1
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
	address := fmt.Sprintf("%s:%s", member.Address, member.Port)
	resp, err := sender.SendCommand(&payload, address)
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
