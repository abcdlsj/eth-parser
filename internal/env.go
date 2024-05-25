package internal

import (
	"fmt"
	"os"
)

var (
	RELAY_FILE = orEnv("RELAY_FILE", "testdata/relay.json")

	RELAY_FLAG = orEnv("RELAY", "false")
	MOCK_FLAG  = orEnv("MOCK", "false")

	PORT = orEnv("PORT", "8080")

	ServerURL = orEnv("SERVER_URL", fmt.Sprintf("http://localhost:%s", PORT))
)

func orEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
