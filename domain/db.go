package domain

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

// const PATH_TO_DB = "~/executor.db"
const PATH_TO_DB = "/tmp/executor.db"

const TasksTabelDefinition = `
CREATE TABLE IF NOT EXISTS tasks(
id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
lim INTEGER NOT NULL,
type INTEGER NOT NULL,
status INTEGER NOT NULL,
output TEXT,
cmd TEXT NOT NULL,
time_start TEXT,
time_finish TEXT
);`

type SQLiteRepository struct {
	db *sql.DB
}

func createDB() *sql.DB {
	db, err := sql.Open("sqlite3", PATH_TO_DB)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(TasksTabelDefinition)
	if err != nil {
		fmt.Errorf("Error: %s", err)
		os.Exit(-1)
	}
	return db
}

func NewSQLiteRepository() *SQLiteRepository {
	var db *sql.DB

	if _, err := os.Stat(PATH_TO_DB); os.IsNotExist(err) {
		db = createDB()
		fmt.Printf("DB isn't exist. Check: %s\n", PATH_TO_DB)
	} else {
		db, err = sql.Open("sqlite3", PATH_TO_DB)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("DB already exist: %s\n", PATH_TO_DB)
	}

	return &SQLiteRepository{
		db: db,
	}
}

func (s *SQLiteRepository) Close() {
	s.db.Close()
}
