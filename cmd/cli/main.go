package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/abcdlsj/eth-parser/internal"
)

func main() {
	usage := func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("  getCurrentBlock\n")
		fmt.Printf("  subscribe <ADDRESS>\n")
		fmt.Printf("  getTransactions <ADDRESS>\n")
	}

	if len(os.Args) < 2 || os.Args[1] == "-h" ||
		((os.Args[1] == "subscribe" || os.Args[1] == "getTransactions") && len(os.Args) < 3) {
		usage()
		return
	}

	switch os.Args[1] {
	case "getCurrentBlock":
		getCurrentBlock()
	case "subscribe":
		address := os.Args[2]
		subscribe(address)
	case "getTransactions":
		address := os.Args[2]
		getTransactions(address)
	default:
		usage()
		return
	}
}

func getCurrentBlock() {
	resp, err := http.Get(fmt.Sprintf("%s/getCurrentBlock", internal.ServerURL))
	if err != nil {
		fmt.Printf("Error getting current block: %v\n", err)
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Println(string(body))
}

func subscribe(address string) {
	data := map[string]string{"address": address}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshalling data: %v\n", err)
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/subscribe", internal.ServerURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error subscribing to address: %v\n", err)
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Println(string(body))
}

func getTransactions(address string) {
	resp, err := http.Get(fmt.Sprintf("%s/getTransactions?address=%s", internal.ServerURL, address))
	if err != nil {
		fmt.Printf("Error getting transactions for address: %v\n", err)
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Println(string(body))
}
