package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ab-testing-service/internal/models"
)

func (s *Server) updateProxyCondition(c *gin.Context) {
	proxyID := c.Param("id")
	var condition models.RouteCondition
	if err := c.ShouldBindJSON(&condition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	var userID *string
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			userID = &u.ID
		}
	}

	// Update condition in storage
	if err := s.storage.UpdateProxyCondition(c.Request.Context(), proxyID, &condition, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
