BUILDPATH=$(CURDIR)
GO=$(shell which go)
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOMOD=$(GO) mod init
GOTEST=$(GO) test -v ./...

RAFTSERVER=cmd/raftserver
RAFTSERVER_PATH=$(BUILDPATH)/$(RAFTSERVER)/


server:
	cd $(RAFTSERVER_PATH);$(GOBUILD)

server-clean:
	cd $(RAFTSERVER_PATH);$(GOCLEAN)

all: server

clean-all: server-clean


