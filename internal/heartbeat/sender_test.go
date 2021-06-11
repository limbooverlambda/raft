package heartbeat

import (
	"github.com/kitengo/raft/internal/member"
	"github.com/kitengo/raft/internal/models"
	raftstate "github.com/kitengo/raft/internal/state"
	"testing"
)

func Test_raftHeartbeat_SendHeartbeats(t *testing.T) {
	fakeTerm := fakeRaftTerm{}
	fakeTerm.GetTermFn = func() uint64 {
		return 1
	}
	fakeMember := fakeRaftMember{}
	fakeMember.GetLeaderFn = func() member.Entry {
		return member.Entry{
			ID:      "leader",
			Address: "127.0.0.1",
			Port:    "21",
		}
	}
	fakeMember.GetListFn = func() ([]member.Entry, error) {
		return []member.Entry{
			{
				ID:      "peer1",
				Address: "127.0.0.1",
				Port:    "23",
			},
			{
				ID:      "peer2",
				Address: "127.0.0.1",
				Port:    "24",
			},
		}, nil
	}
	raftIndex := raftstate.NewRaftIndex()
	fakeSender := fakeRaftSender{}
	expectedSendCount := 2
	var actualSendCount int
	fakeSender.GetSendCommandFn = func(requestConv models.RequestConverter, ip, port string) (response models.Response, err error) {
		actualSendCount++
		return models.Response{}, nil
	}
	hb := NewRaftHeartbeat(fakeTerm,
		fakeMember, raftIndex, fakeSender)
	hb.SendHeartbeats()
	if actualSendCount != expectedSendCount {
		t.Errorf("expected %v but actual %v", expectedSendCount, actualSendCount)
	}
}
