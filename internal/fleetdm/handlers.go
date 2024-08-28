package fleetdm

import (
	"encoding/json"
	"net/http"
)

func (srv *server) getIndexHandler(w http.ResponseWriter, r *http.Request) {
	records, err := srv.getRecords(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
