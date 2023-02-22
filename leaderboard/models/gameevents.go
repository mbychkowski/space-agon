package models

import (
	"gorm.io/gorm"
	"cloud.google.com/go/spanner"
	_ "github.com/googleapis/go-sql-spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
)

type GameEvent struct {
	EventID   string
	PlayerID  int64
	EventType string
	Timestamp int64
	Data      interface{}
}

const (
	PROJECT  = "prj-zeld-gke"
	INSTANCE = "spaceagon-demo"
	DATABASE = "spaceagon-db-demo"
)

var DB *gorm.DB

func ConnectDatabase() {

  spannerDB := "projects/"+PROJECT+"/instances/"+INSTANCE+"/databases/"+DATABASE

	sqlDB, err := sql.Open("spanner", spannerDB)
	if err != nil {
		log.Fatal(err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
  	Conn: sqlDB,
	}), &gorm.Config{})

	err = gormDB.AutoMigrate(&GameEvent{})
	if err != nil {
		return
	}

	DB = gormDB
}
