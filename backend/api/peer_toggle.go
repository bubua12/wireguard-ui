package api

import (
	"net/http"
	"strconv"
	"wireguard-ui/db"

	"github.com/gin-gonic/gin"
)

type TogglePeerReq struct {
	Enabled bool `json:"enabled"`
}

func TogglePeer(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req TogglePeerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.TogglePeer(id, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Toggled", "enabled": req.Enabled})
}
