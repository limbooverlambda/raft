BUILDPATH=$(CURDIR)
GO=$(shell which go)
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOMOD=$(GO) mod init
GOTEST=$(GO) test -v ./...
GOFMT=$(GO) fmt ./...

RAFTSERVER=cmd/raftserver
RAFTSERVER_PATH=$(BUILDPATH)/$(RAFTSERVER)/

RAFTCLIENT=cmd/raftclient
RAFTCLIENT_PATH=$(BUILDPATH)/$(RAFTCLIENT)/

LOG_FILE=log*

test:
	$(GOTEST)

server:
	cd $(RAFTSERVER_PATH);$(GOFMT);$(GOBUILD)

server-clean:
	cd $(RAFTSERVER_PATH);$(GOCLEAN);rm -f $(LOG_FILE)

client:
	cd $(RAFTCLIENT_PATH);$(GOFMT);$(GOBUILD)

client-clean:
	cd $(RAFTCLIENT_PATH);$(GOCLEAN)


all: server client test

clean-all: server-clean client-clean


