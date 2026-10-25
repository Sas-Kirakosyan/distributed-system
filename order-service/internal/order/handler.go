package order

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// dbTimeout bounds every DB call made from a request, so a slow or dead
// database can't hang a request forever.
const dbTimeout = 3 * time.Second

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/orders", h.create)
	r.GET("/orders/:id", h.get)
}

type createOrderRequest struct {
	UserID string `json:"user_id" binding:"required,max=64"`
	Amount int64  `json:"amount" binding:"required,gt=0"`
}

// POST /orders  {"user_id": "u1", "amount": 1999}
func (h *Handler) create(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), dbTimeout)
	defer cancel()

	o, err := h.repo.Create(ctx, req.UserID, req.Amount)
	if err != nil {
		internalError(c, "create order", err)
		return
	}
	c.JSON(http.StatusCreated, o)
}

// GET /orders/:id
func (h *Handler) get(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), dbTimeout)
	defer cancel()

	o, err := h.repo.GetByID(ctx, id)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		internalError(c, "get order", err)
		return
	}
	c.JSON(http.StatusOK, o)
}

// internalError logs the real error and returns a generic 500, so DB
// details never leak to clients.
func internalError(c *gin.Context, op string, err error) {
	slog.Error(op+" failed", "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
