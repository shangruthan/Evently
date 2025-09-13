package handler

import (
	"encoding/json"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(map[string]interface{}{"data": payload, "error": nil})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RespondWithError(w http.ResponseWriter, code int, errCode, message string) {
	errPayload := map[string]string{"code": errCode, "message": message}
	response, _ := json.Marshal(map[string]interface{}{"data": nil, "error": errPayload})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
