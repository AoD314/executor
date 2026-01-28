package domain

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

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

func createDB(path_to_db string) *sql.DB {
	db, err := sql.Open("sqlite3", path_to_db)
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	path_to_db := filepath.Join(homeDir, "executor.db")

	if _, err := os.Stat(path_to_db); os.IsNotExist(err) {
		db = createDB(path_to_db)
		fmt.Printf("DB isn't exist. Check: %s\n", path_to_db)
	} else {
		db, err = sql.Open("sqlite3", path_to_db)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("DB already exist: %s\n", path_to_db)

		// fix after crash

		q := fmt.Sprintf("UPDATE tasks SET status = %d WHERE status = %d;", STATUS_TASK_WAITING, STATUS_TASK_RUNNING)
		fmt.Println(q)

		_, err := db.Exec(q)
		if err != nil {
			fmt.Errorf("Error: %s", err)
		}
	}

	return &SQLiteRepository{
		db: db,
	}
}

func (s *SQLiteRepository) Close() {
	s.db.Close()
}
