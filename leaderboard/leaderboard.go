package main

import (
  "github.com/gin-gonic/gin"
	"github.com/mbychkowski/space-agon/leaderboard/controllers"
)

func main() {
  r := gin.Default()

	r.GET("/events", controllers.GetEvents)

	r.GET("/leaderboard", controllers.GetLeaderboard)

  r.Run()
}
