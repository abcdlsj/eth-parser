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

test: daemon cli
	MOCK=true ./$(DAEMON) 2>&1 /dev/null &

	# run cli
	./$(CLI) subscribe -address "0xae2fc483527b8ef99eb5d9b44875f005ba1fae13"
	# time sleep 11s # wait for getTransactions
	sleep 11
	./$(CLI) getTransactions -address "0xae2fc483527b8ef99eb5d9b44875f005ba1fae13"

	# kill daemon
	pkill $(DAEMON)