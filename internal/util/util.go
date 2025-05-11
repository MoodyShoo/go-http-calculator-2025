package util

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/MoodyShoo/go-http-calculator/internal/models"
)

const (
	ContentType     = "Content-Type"
	ApplicationJson = "application/json"
)

// sendResponse отправляет ответ клиенту
func SendResponse(w http.ResponseWriter, response models.Response, status int) {
	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(status)
	resp, err := response.ToJSON()
	if err != nil {
		SendError(w, "Failed to encode response", status)
		return
	}
	log.Printf("Response sent.")
	w.Write(resp)
}

// sendError отправляет ошибку клиенту.
func SendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(status)
	resp, _ := json.Marshal(map[string]string{"error": message})
	log.Printf("Error response sent.")
	w.Write(resp)
}
