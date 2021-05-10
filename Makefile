BUILDPATH=$(CURDIR)
GO=$(shell which go)
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOMOD=$(GO) mod init
GOTEST=$(GO) test -v ./...

RAFTSERVER=cmd/raftserver
RAFTSERVER_PATH=$(BUILDPATH)/$(RAFTSERVER)/

RAFTCLIENT=cmd/raftclient
RAFTCLIENT_PATH=$(BUILDPATH)/$(RAFTCLIENT)/

LOG_FILE=log
LOGIDX_FILE=log_idx

server:
	cd $(RAFTSERVER_PATH);$(GOBUILD)

server-clean:
	cd $(RAFTSERVER_PATH);$(GOCLEAN);rm $(LOG_FILE);rm $(LOGIDX_FILE)

client:
	cd $(RAFTCLIENT_PATH);$(GOBUILD)

client-clean:
	cd $(RAFTCLIENT_PATH);$(GOCLEAN)


all: server client

clean-all: server-clean client-clean


