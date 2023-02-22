// controllers/books.go

package controllers

import (
	"context"
	"fmt"
	"net/http"

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

	sqlStr := `SELECT PlayerId, count(EventType) FROM gameevents WHERE EventType="SpawnMissile" GROUP by PlayerId`

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
			// return nil
			break
		}
		if err != nil {
			// return err
			fmt.Println("error retrieving Spanner results:", err)
		}
		// if err := row.Columns(&eventId, &playerId, &eventType, &timestamp, &data); err != nil {
		if err := row.Columns(&playerId, &count); err != nil {
			// return err
			fmt.Println("error parsing Spanner result:", err)
		}

		players = append(players, models.Player{
			Name: playerId,
			Score: count,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": players})
}
