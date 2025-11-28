package controllers

import (
	"net/http"
	"pariwisata/config"
	"pariwisata/models"

	"github.com/gin-gonic/gin"
)

func ApiGetDestinations(c *gin.Context) {
	var destinations []models.Destination
	// Preload relasi untuk data lengkap JSON
	config.DB.Preload("Category").Preload("Galleries").Find(&destinations)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   destinations,
	})
}