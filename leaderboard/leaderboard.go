package main

import (
	"net/http"

  "github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/mbychkowski/space-agon/leaderboard/controllers"
)

func main() {
  r := gin.Default()
	r.Use(cors.Default())

	r.GET("/events", controllers.GetEvents)

	r.GET("/leaderboard", controllers.GetLeaderboard)

	r.GET("/healthz", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"data": "ok"})
  })

  r.Run()
}
