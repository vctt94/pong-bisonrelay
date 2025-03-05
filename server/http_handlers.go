package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
)

// handleFetchTipsByClientIDHandler fetches tips for a specific client ID.
func (s *Server) handleFetchTipsByClientIDHandler(w http.ResponseWriter, r *http.Request) {
	clientIDStr := r.URL.Query().Get("clientID")
	if clientIDStr == "" {
		http.Error(w, "clientID parameter is required", http.StatusBadRequest)
		return
	}

	var clientID zkidentity.ShortID
	if err := clientID.FromString(clientIDStr); err != nil {
		http.Error(w, fmt.Sprintf("invalid client ID: %v", err), http.StatusBadRequest)
		return
	}

	tips, err := s.db.FetchAllReceivedTipsByUID(context.Background(), clientID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching tips: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tips)
}

// handleFetchAllUnprocessedTipsHandler fetches all unprocessed tips for all clients.
func (s *Server) handleFetchAllUnprocessedTipsHandler(w http.ResponseWriter, r *http.Request) {
	tips, err := s.db.FetchUnprocessedTips(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching unprocessed tips: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert the map[zkidentity.ShortID][]serverdb.ReceivedTipWrapper to map[string][]serverdb.ReceivedTipWrapper
	response := make(map[string][]*types.ReceivedTip)
	for clientID, clientTips := range tips {
		response[clientID.String()] = clientTips // Convert clientID to string
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("error encoding response: %v", err), http.StatusInternalServerError)
	}
}

// handleGetSendProgressByWinnerHandler retrieves progress for a specific winner
func (s *Server) handleGetSendProgressByWinnerHandler(w http.ResponseWriter, r *http.Request) {
	clientIDStr := r.URL.Query().Get("clientID")
	if clientIDStr == "" {
		http.Error(w, "clientID parameter is required", http.StatusBadRequest)
		return
	}

	var clientID zkidentity.ShortID
	if err := clientID.FromString(clientIDStr); err != nil {
		http.Error(w, fmt.Sprintf("invalid client ID: %v", err), http.StatusBadRequest)
		return
	}

	records, err := s.db.FetchSendTipProgressByClient(r.Context(), clientID.Bytes())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching progress: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}
