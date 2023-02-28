// controllers/books.go

package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/mbychkowski/space-agon/leaderboard/models"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func GetEvents(c *gin.Context) {
	var gameEvents []models.GameEvent

	ctx := context.Background()
	client := models.ConnectDatabase(ctx)
	defer client.Close()

	sqlStr := `SELECT PlayerId, EventId, EventType, Timestamp, Data FROM gameevents`

	statement := spanner.Statement{
		SQL: sqlStr,
	}

	// Execute the query and retrieve the results
	iter := client.Single().Query(ctx, statement)
	defer iter.Stop()

	for {
		var eventId 	string
		var playerId 	int64
		var eventType string
		var timestamp int64
		var data 			interface{}

		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("error retrieving Spanner results:", err)
		}
		if err := row.Columns(&eventId, &playerId, &eventType, &timestamp, &data); err != nil {
			fmt.Println("error parsing Spanner result:", err)
		}

		gameEvents = append(gameEvents, models.GameEvent{
			EventId: eventId,
			PlayerId: playerId,
			EventType: eventType,
			Timestamp: timestamp,
			Data: data,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": gameEvents})
}

func GetLeaderboard(c *gin.Context) {
	var players []models.Player

	ctx := context.Background()
	client := models.ConnectDatabase(ctx)
	defer client.Close()

	sqlStr := `SELECT PlayerId, COUNT(EventType) FROM gameevents WHERE EventType="SpawnMissile" GROUP by PlayerId ORDER BY COUNT(EventType) DESC`

	statement := spanner.Statement{
		SQL: sqlStr,
	}

	// Execute the query and retrieve the results
	iter := client.Single().Query(ctx, statement)
	defer iter.Stop()

	for {
		var playerId 	string
		var count 		int64

		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("error retrieving Spanner results:", err)
		}
		if err := row.Columns(&playerId, &count); err != nil {
			fmt.Println("error parsing Spanner result:", err)
		}

		players = append(players, models.Player{
			Name: playerId,
			Score: count,
		})
	}

	jsonPlayers, err := json.Marshal(&players)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Players Marshalled:", string(jsonPlayers))

	c.IndentedJSON(http.StatusOK, gin.H{
		"message": string(jsonPlayers),
	})
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
}
