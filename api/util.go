package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// respondWithError sets the given status code and
// transforms error into a JSON format.
func respondWithError(w http.ResponseWriter, code int, err error) {
	log.Printf("[API]::ERROR:: %v\n", err)
	respondWithJSON(w, code, map[string]string{"error": err.Error()})
}

// respondWithJSON marshals the payload in the body of the response,
// sets the content-type and and writes the given status code (2XX).
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
