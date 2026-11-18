package v2

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func fetchURL(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty url"})
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// VULN(SSRF): arbitrary user-supplied URL is fetched server-side.
	resp, err := client.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var data any
	if json.Unmarshal(body, &data) != nil {
		data = string(body)
	}

	c.JSON(http.StatusOK, gin.H{
		"url":  url,
		"data": data,
	})
}
