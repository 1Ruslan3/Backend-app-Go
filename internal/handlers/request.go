package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"premium_cars_app/internal/queue"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gorilla/mux"
)

type RequestHandler struct {
	Queue *queue.Queue
}

func (h *RequestHandler) CreateRequestGin(c *gin.Context) {
	type requestBody struct {
		CarModel string `json:"car_model"`
	}

	var body requestBody
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.CarModel) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "car_model is required"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	requestTime := c.GetString("requestTime")

	req := &queue.Request{
		ID:        uuid.NewString(),
		UserID:    userID.(string),
		CarModel:  strings.TrimSpace(body.CarModel),
		CreatedAt: requestTime,
	}

	if err := h.Queue.Enqueue(context.Background(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"request_id": req.ID,
		"status":     "queued",
	})
}

func (h *RequestHandler) GetRequestStatusGin(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request id required"})
		return
	}

	status, err := h.Queue.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"request_id": id,
		"status":     status,
	})
}

func (h *RequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		CarModel string `json:"car_model"`
	}

	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	body.CarModel = strings.TrimSpace(body.CarModel)
	if body.CarModel == "" {
		http.Error(w, "car_model is required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req := &queue.Request{
		ID:        uuid.NewString(),
		UserID:    userID.(string),
		CarModel:  body.CarModel,
		CreatedAt: r.Context().Value("requestTime").(string),
	}

	if err := h.Queue.Enqueue(context.Background(), req); err != nil {
		http.Error(w, "Failed to enqueue", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"request_id": req.ID,
		"status":     "queued",
	})
}

func (h *RequestHandler) GetRequestStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing request id", http.StatusBadRequest)
		return
	}

	status, err := h.Queue.GetStatus(r.Context(), id)
	if err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"request_id": id,
		"status":     status,
	})
}
