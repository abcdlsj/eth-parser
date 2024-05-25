DAEMON=bin/ethparser-server
CLI=bin/ethparser-cli

GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean

DAEMON_SRC=cmd/server/main.go
CLI_SRC=cmd/cli/main.go

.PHONY: all clean daemon cli relay

all: daemon cli

daemon:
	$(GOBUILD) -o $(DAEMON) $(DAEMON_SRC)

cli:
	$(GOBUILD) -o $(CLI) $(CLI_SRC)

clean:
	$(GOCLEAN)
	rm -f $(DAEMON) $(CLI)

relay: daemon
	RELAY=true ./$(DAEMON)

mock: daemon
	MOCK=true ./$(DAEMON)

# this is for myown testing, don't use, need to record new transactions
test: daemon cli
	MOCK=true ./$(DAEMON) 2>&1 > /dev/null &
	sleep 1

	RELAY_FILE=testdata/relay.json ./$(CLI) subscribe 0xa7efae728d2936e78bda97dc267687568dd593f3
	sleep 20
	./$(CLI) getTransactions 0xa7efae728d2936e78bda97dc267687568dd593f3

	pkill ethparser-server