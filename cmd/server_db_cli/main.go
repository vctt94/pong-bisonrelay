package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	serverURL = flag.String("serverURL", "http://localhost:8888", "URL of the HTTP server")
)

type Tip struct {
	Tip struct {
		Uid          string `json:"uid"`
		AmountMatoms int64  `json:"amount_matoms"`
		SequenceId   uint64 `json:"sequence_id"`
	} `json:"Tip"`
	Status string `json:"Status"`
}

type TipProgressRecord struct {
	SequenceID  uint64    `json:"id"`
	WinnerUID   string    `json:"winner_uid"`
	TotalAmount int64     `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Available commands:")
		fmt.Println("  getsendprogress <clientID>")
		fmt.Println("  getreceived <clientID>")
		fmt.Println("  getall")
		os.Exit(1)
	}

	switch args[0] {
	case "getsendprogress":
		if len(args) < 2 {
			fmt.Println("Client ID required for getsendprogress")
			os.Exit(1)
		}
		fetchSendProgress(args[1])
	case "getreceived":
		if len(args) < 2 {
			fmt.Println("Client ID required for getreceived")
			os.Exit(1)
		}
		fetchReceivedTips(args[1])
	case "getall":
		fetchAllUnprocessedTips()
	default:
		fmt.Println("Unknown command:", args[0])
		os.Exit(1)
	}
}

func fetchReceivedTips(clientID string) {
	url := fmt.Sprintf("%s/received?clientID=%s", *serverURL, clientID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching received tips: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server returned error: %s\n", resp.Status)
		return
	}

	var tips []Tip
	if err := json.NewDecoder(resp.Body).Decode(&tips); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	fmt.Printf("Received tips for client ID %s:\n", clientID)
	for _, tip := range tips {
		fmt.Printf("- Tip ID: %s\n  Amount: %d matoms\n  Status: %s\n",
			tip.Tip.Uid, tip.Tip.AmountMatoms, tip.Status)
	}
}

func fetchAllUnprocessedTips() {
	url := fmt.Sprintf("%s/fetchAllUnprocessedTips", *serverURL)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching all unprocessed tips: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check if the response status is OK (200)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server returned error: %s\n", resp.Status)
		return
	}

	// Decode response body
	var tips map[string][]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tips); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	// Display the unprocessed tips
	fmt.Println("Unprocessed Tips:")
	if len(tips) == 0 {
		fmt.Println("No unprocessed tips found.")
	} else {
		for clientID, clientTips := range tips {
			fmt.Printf("Client %s:\n", clientID)
			for _, tip := range clientTips {
				fmt.Printf("Tip: %+v\n", tip)
			}
		}
	}
}

func fetchSendProgress(clientID string) {
	url := fmt.Sprintf("%s/tipprogress?clientID=%s", *serverURL, clientID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching send progress: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server returned error: %s\n", resp.Status)
		return
	}

	var records []struct {
		CreatedAt    time.Time `json:"created_at"`
		ID           int       `json:"id"`
		ReceivedTips []struct {
			AmountMatoms int64  `json:"amount_matoms"`
			SequenceID   int64  `json:"sequence_id"`
			UID          string `json:"uid"`
		} `json:"received_tip"`
		Status      string `json:"status"`
		TotalAmount int64  `json:"total_amount"`
		WinnerUID   string `json:"winner_uid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSend Progress Records for client %s:\n\n", clientID)
	for i, record := range records {
		fmt.Printf("Record #%d\n", i+1)
		fmt.Printf("• Created:   %s\n", record.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("• Status:    %s\n", strings.ToUpper(record.Status))
		fmt.Printf("• Total:     %s DCR\n", formatMatoms(record.TotalAmount))
		fmt.Printf("• Winner:    %s\n", truncateUID(record.WinnerUID))
		fmt.Println("• Received Tips:")

		for _, tip := range record.ReceivedTips {
			fmt.Printf("  - ID: %-12d | %8s DCR | UID: %s\n",
				tip.SequenceID,
				formatMatoms(tip.AmountMatoms),
				truncateUID(tip.UID),
			)
		}
		fmt.Println()
	}
}

func formatMatoms(matoms int64) string {
	btc := float64(matoms) / 1e11 // Convert matoms to BTC
	return fmt.Sprintf("%.2f", btc)
}

func truncateUID(uid string) string {
	if len(uid) <= 12 {
		return uid
	}
	return uid[:6] + "..." + uid[len(uid)-6:]
}
