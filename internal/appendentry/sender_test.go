package appendentry

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/kitengo/raft/internal/models"
	"testing"
	"time"
)

func TestNewSender_HappyPath(t *testing.T) {
	expectedResponse := models.AppendEntryResponse{
		Term:    1,
		Success: true,
	}
	fLog := fakeLog{}
	fSender := fakeSender{
		GetSendCommandFn: func(requestConv models.RequestConverter, ip,
			port string) (response models.Response, err error) {
			aer := expectedResponse
			var payloadBytes bytes.Buffer
			enc := gob.NewEncoder(&payloadBytes)
			err = enc.Encode(aer)
			return models.Response{Payload: payloadBytes.Bytes()}, err
		}}

	ctxt := context.TODO()
	sender := NewSender(ctxt, fSender, fLog)
	respChan := make(chan models.AppendEntryResponse, 1)
	entry := Entry{
		RespChan: respChan,
	}
	sender.ForwardEntry(entry)
	actualResponse := <-respChan
	if actualResponse != expectedResponse {
		t.Errorf("expected %v but actual %v", expectedResponse, actualResponse)
	}
}

func TestNewSender_WithSendFailureRetries(t *testing.T) {
	var sendCount int
	fLog := fakeLog{}
	fSender := fakeSender{
		GetSendCommandFn: func(requestConv models.RequestConverter, ip,
			port string) (response models.Response, err error) {
			sendCount++
			return models.Response{}, err
		}}

	ctxt, cancel := context.WithCancel(context.Background())
	sender := NewSender(ctxt, fSender, fLog)
	respChan := make(chan models.AppendEntryResponse, 1)
	entry := Entry{
		RespChan: respChan,
	}
	sender.ForwardEntry(entry)
	time.AfterFunc(2*time.Second, cancel)
	<-respChan
	if sendCount <= 1 {
		t.Errorf("expected resends but actual send count %v", sendCount)
	}
}
