package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"premium_cars_app/internal/queue"
	"premium_cars_app/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	userID, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req := &models.Request{
		ID:        uuid.New().String(),
		UserID:    userID.(string),
		CarModel:  body.CarModel,
		Status:    "queued",
		CreatedAt: time.Now(),
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
