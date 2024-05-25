package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/abcdlsj/eth-parser/internal"
)

func main() {
	parser := internal.NewEthParser()

	var wg sync.WaitGroup

	srv := serverRouter(parser)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		parser.Run()
		fmt.Println("parser released")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Eth parser server is running on port " + internal.PORT)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
		fmt.Println("server released")
	}()

	<-sig
	log.Println("Received signal, shutting down...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("Shutdown server error: %v", err)
	}

	parser.Stop()
	if internal.RELAY_FILE == "true" {
		parser.SaveRelay()
	}

	wg.Wait()
	log.Println("Shutdown complete")
}

func serverRouter(p *internal.EthParser) *http.Server {
	srv := http.Server{
		Addr: ":" + internal.PORT,
	}

	http.HandleFunc("/getCurrentBlock", func(w http.ResponseWriter, r *http.Request) {
		blockNumber := p.GetCurrentBlock()
		if err := json.NewEncoder(w).Encode(map[string]int{"current_block": blockNumber}); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Address string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		success := p.Subscribe(req.Address)
		if err := json.NewEncoder(w).Encode(map[string]bool{"subscribed": success}); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/getTransactions", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Missing address parameter", http.StatusBadRequest)
			return
		}
		transactions := p.GetTransactions(address)
		if err := json.NewEncoder(w).Encode(transactions); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	if internal.RELAY_FLAG == "true" {
		http.HandleFunc("/saveRelay", func(w http.ResponseWriter, r *http.Request) {
			p.SaveRelay()
		})
	}

	return &srv
}
