package main

import (
  "net/http"
  "github.com/gin-gonic/gin"
	// "github.com:mbychkowski/space-agon/leaderboard/models"
)

func main() {
  r := gin.Default()

  r.GET("/", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"data": "hello world"})
  })

  r.Run()
}
