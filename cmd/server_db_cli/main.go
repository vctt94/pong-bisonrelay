package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	serverURL = flag.String("serverURL", "http://localhost:8888", "URL of the HTTP server")
	clientID  = flag.String("client", "", "Client ID to fetch tips for")
	showAll   = flag.Bool("all", false, "Show all unprocessed tips")
)

type Tip struct {
	Tip struct {
		Uid          string `json:"uid"`
		AmountMatoms int64  `json:"amount_matoms"`
		SequenceId   uint64 `json:"sequence_id"`
	} `json:"Tip"`
	Status string `json:"Status"`
}

func main() {
	flag.Parse()

	if *clientID != "" {
		fetchTipsByClientID(*clientID)
	} else if *showAll {
		fetchAllUnprocessedTips()
	} else {
		fmt.Println("Please provide either a client ID (--clientID) or use --all to display all unprocessed tips.")
		flag.Usage()
		os.Exit(1)
	}
}

func fetchTipsByClientID(clientID string) {
	url := fmt.Sprintf("%s/fetchTipsByClientID?clientID=%s", *serverURL, clientID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching tips by client ID: %v\n", err)
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

	fmt.Printf("Tips for client ID %s:\n", clientID)
	fmt.Printf("%+v\n\n", tips)
	for _, tip := range tips {
		fmt.Println(tip)
		// fmt.Printf("tip amt: %+v\n\n", tip.Tip.AmountMatoms)
		// fmt.Printf("Tip ID: %x\nAmount (matoms): %.8f\nStatus: %v\nTimestamp: %v\n\n",
		// 	tip.Tip.SequenceId, float64(tip.Tip.AmountMatoms)/1e11, tip.Status, time.UnixMilli(tip.Tip.TimestampMs))
	}
	// for _, tip := range tips {

	// 	fmt.Printf("Tip: %+v\n", tip)
	// }
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
