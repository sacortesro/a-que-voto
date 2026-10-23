package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithInternalError(w http.ResponseWriter, op string, err error) {
	log.Printf("%s: %v", op, err)
	RespondWithError(w, http.StatusInternalServerError, "internal server error")
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	if code >= http.StatusInternalServerError {
		log.Println("Responding with 5xx error:", msg)
	}

	type errResponse struct {
		Error string `json:"error"`
	}

	RespondWithJSON(w, code, errResponse{Error: msg})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
