package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

const (
	ADD    = 1
	REMOVE = 2
	EDIT   = 3
	PRINT  = 4
	QUIT   = 5
)

type Word struct {
	Word       string    `json:"word"`
	Definition string    `json:"definition"`
	Book       string    `json:"book"`
	Author     string    `json:"author"`
	Language   string    `json:"language"`
	Created_at time.Time `json:"created_at"`
}

var dbconn *pgx.Conn

func main() {
	connected := connect()
	defer dbconn.Close(context.Background())
	if !connected {
		return
	}
	router := gin.Default()
	router.POST("/add", add_word)
	router.Run()
}

func add_word(c *gin.Context) {
	var word Word
	err := c.ShouldBind(&word)
	if err != nil {
		log.Fatalf("Couldn't add word: %v", err)
		return
	}
	_, err = dbconn.Exec(context.Background(),
		"INSERT INTO words (word, definition, created_at, book, author, language) VALUES ($1, $2,$3,$4,$5,$6)",
		word.Word, word.Definition, word.Created_at, word.Book, word.Author, word.Language)
	if err != nil {
		log.Fatalf("Failed to add row: %v", err)
	}
	fmt.Println("Word added succesfully")
}

func connect() bool {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		conn.Close(context.Background())
		return false
	}
	// this tests the connection
	var version string
	if err := conn.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		log.Fatalf("Initial Query failed: %v", err)
		conn.Close(context.Background())
		return false
	}
	log.Println("Connected to:", version)
	dbconn = conn
	return true
}
