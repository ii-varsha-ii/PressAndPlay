package common

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func RespondWithError(w http.ResponseWriter, code int, message string) {
	log.Errorf("respondWithError: %v", message)
	RespondWithJSON(w, code, "", map[string]string{"error": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, sessionID string, payload interface{}) {
	response, _ := json.Marshal(payload)

	if len(sessionID) > 0 {
		w.Header().Set("User-Session-Id", sessionID)
	}

	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RespondWithStatusCode(w http.ResponseWriter, code int, headers map[string]string) {
	if len(headers) > 0 {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
	}
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(code)
}
