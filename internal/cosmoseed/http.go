package cosmoseed

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Seeder) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/peers", s.handlePeers)
}

func (s *Seeder) handlePeers(w http.ResponseWriter, r *http.Request) {
	peers := s.pex.GetPeerSelection()

	peerList := make([]string, 0, len(peers))
	for _, p := range peers {
		peerList = append(peerList, p.String())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strings.Join(peerList, ","))
}
