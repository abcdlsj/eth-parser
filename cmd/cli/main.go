package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

var serverURL = "http://localhost:" + orEnv("PORT", "8080")

func main() {
	getCurrentBlockCmd := flag.NewFlagSet("getCurrentBlock", flag.ExitOnError)
	subscribeCmd := flag.NewFlagSet("subscribe", flag.ExitOnError)
	getTransactionsCmd := flag.NewFlagSet("getTransactions", flag.ExitOnError)

	subscribeAddress := subscribeCmd.String("address", "", "The address to subscribe to")
	getTransactionsAddress := getTransactionsCmd.String("address", "", "The address to get transactions for")

	flag.StringVar(&serverURL, "server", serverURL, "The server URL")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'getCurrentBlock', 'subscribe' or 'getTransactions' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "getCurrentBlock":
		getCurrentBlockCmd.Parse(os.Args[2:])
		getCurrentBlock()
	case "subscribe":
		subscribeCmd.Parse(os.Args[2:])
		if *subscribeAddress == "" {
			fmt.Println("Please provide an address to subscribe")
			subscribeCmd.PrintDefaults()
			os.Exit(1)
		}
		subscribe(*subscribeAddress)
	case "getTransactions":
		getTransactionsCmd.Parse(os.Args[2:])
		if *getTransactionsAddress == "" {
			fmt.Println("Please provide an address to get transactions for")
			getTransactionsCmd.PrintDefaults()
			os.Exit(1)
		}
		getTransactions(*getTransactionsAddress)
	default:
		fmt.Println("Expected 'getCurrentBlock', 'subscribe' or 'getTransactions' subcommands")
		os.Exit(1)
	}
}

func getCurrentBlock() {
	resp, err := http.Get(fmt.Sprintf("%s/getCurrentBlock", serverURL))
	if err != nil {
		fmt.Printf("Error getting current block: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(body))
}

func subscribe(address string) {
	data := map[string]string{"address": address}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshalling data: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(fmt.Sprintf("%s/subscribe", serverURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error subscribing to address: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(body))
}

func getTransactions(address string) {
	resp, err := http.Get(fmt.Sprintf("%s/getTransactions?address=%s", serverURL, address))
	if err != nil {
		fmt.Printf("Error getting transactions for address: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(body))
}

func orEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
