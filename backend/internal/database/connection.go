//file: backend/db/database.go

package db

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func Init() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Erreur de connexion à la base : %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Base de données inaccessible : %v", err)
	}
}

