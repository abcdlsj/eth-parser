package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/abcdlsj/eth-parser/internal"
)

func main() {
	parser := internal.NewEthParser()

	go parser.Run()
	defer parser.Stop()

	http.HandleFunc("/getCurrentBlock", func(w http.ResponseWriter, r *http.Request) {
		blockNumber := parser.GetCurrentBlock()
		json.NewEncoder(w).Encode(map[string]int{"current_block": blockNumber})
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Address string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		success := parser.Subscribe(req.Address)
		json.NewEncoder(w).Encode(map[string]bool{"subscribed": success})
	})

	http.HandleFunc("/getTransactions", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Missing address parameter", http.StatusBadRequest)
			return
		}
		transactions := parser.GetTransactions(address)
		json.NewEncoder(w).Encode(transactions)
	})

	fmt.Println("Eth parser server is running on port 8080")
	http.ListenAndServe(":"+orEnv("PORT", "8080"), nil)
}

func orEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
