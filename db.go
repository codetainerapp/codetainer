package codetainer

import (
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-xorm/xorm"
)

var (
	//
	// Default command to start in a container
	//
	DefaultExecCommand string = "/bin/bash"
)

type CodetainerImage struct {
	Id             string
	Tags           []string
	DefaultCommand string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Codetainer struct {
	Id        string
	ImageId   string
	Defunct   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Database struct {
	engine *xorm.Engine
}

func CloseDb(db *Database) {
	db.engine.Close()
}

func NewDatabase() (*Database, error) {
	db := &Database{}
	dbPath := GlobalConfig.GetDatabasePath()
	engine, err := xorm.NewEngine("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.engine = engine
	runtime.SetFinalizer(db, CloseDb)
	return db, nil
}
