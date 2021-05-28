package sender

import (
	"encoding/gob"
	"fmt"
	raftmodels "github.com/kitengo/raft/internal/models"
	"log"
	"net"
)

type RequestSender interface {
	SendCommand(requestConv raftmodels.RequestConverter, ip, port string) (response raftmodels.Response, err error)
}

func NewRaftRequestSender() RequestSender {
	return requestSender{}
}

type requestSender struct{}

func (rs requestSender) SendCommand(requestConv raftmodels.RequestConverter, ip, port string) (response raftmodels.Response, err error) {
	defer func() {
		if err != nil {
			log.Printf("Encountered error %v\n", err)
		}
	}()
	ipAddress := fmt.Sprintf("%s:%s", ip, port)
	conn, err := net.Dial("tcp", ipAddress)
	if err != nil {
		err = fmt.Errorf("unable to dial %v due to %v", ipAddress, err)
		return
	}
	defer conn.Close()
	req, err := requestConv.ToRequest()
	if err != nil {
		err = fmt.Errorf("unable to serialize payload due to %v", err)
		return
	}
	connEncoder := gob.NewEncoder(conn)
	err = connEncoder.Encode(req)
	if err != nil {
		err = fmt.Errorf("unable to send request %v", err)
		return
	}
	recvChan := make(chan raftmodels.Response)
	go func() {
		defer close(recvChan)
		connDecoder := gob.NewDecoder(conn)
		resp := raftmodels.Response{}
		err = connDecoder.Decode(&resp)
		if err != nil {
			err = fmt.Errorf("unable to decode response %v", err)
			return
		}
		recvChan <- resp
	}()
	response = <-recvChan
	return
}
