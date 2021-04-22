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


server:
	cd $(RAFTSERVER_PATH);$(GOBUILD)

server-clean:
	cd $(RAFTSERVER_PATH);$(GOCLEAN)

client:
	cd $(RAFTCLIENT_PATH);$(GOBUILD)

client-clean:
	cd $(RAFTCLIENT_PATH);$(GOCLEAN)


all: server client

clean-all: server-clean client-clean


