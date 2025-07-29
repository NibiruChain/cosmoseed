package cosmoseed

import (
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

	w.Write([]byte(strings.Join(peerList, ",")))
}
