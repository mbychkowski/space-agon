package models

import (
	"fmt"
	"context"

	"cloud.google.com/go/spanner"
)

type GameEvent struct {
	EventId   string
	PlayerId  int64
	EventType string
	Timestamp int64
	Data      interface{}
}

type Player struct {
	Name  string 	`json:"name"`
	Score int64		`json:"score"`
}

const (
	GCP_PROJECT_ID		= "prj-zeld-gke"
	SPANNER_INSTANCE 	= "spaceagon-demo"
	SPANNER_DATABASE 	= "spaceagon-db-demo"
)

func ConnectDatabase(ctx context.Context) (client *spanner.Client){

	gcpProjectId    := GCP_PROJECT_ID
	spannerInstance := SPANNER_INSTANCE
	spannerDatabase := SPANNER_DATABASE

	spannerConnStr := fmt.Sprintf("projects/%v/instances/%v/databases/%v", gcpProjectId, spannerInstance, spannerDatabase)

	client, err := spanner.NewClient(ctx, spannerConnStr)
	if err != nil {
		fmt.Println("Error creating spanner client", err)
	}
	fmt.Println("Spanner Client created:", spannerConnStr)

	return client
}
