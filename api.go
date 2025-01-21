package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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

	router.Use(cors.Default())

	router.POST("/add", add_word)
	router.POST("/get_words", get_words)
	router.GET("/get_word/:id", word_id)
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

func get_words(c *gin.Context) {
	sort_by := c.Query("sort_by")

	validColumns := map[string]bool{
		"word":       true,
		"definition": true,
		"book":       true,
		"author":     true,
		"language":   true,
		"created_at": true,
	}
	if !validColumns[sort_by] {
		c.JSON(400, gin.H{"error": "Invalid sort_by value"})
		return
	}
	query := fmt.Sprintf("SELECT word, definition, book, author, language, created_at FROM words ORDER BY %s ASC", sort_by)

	rows, err := dbconn.Query(context.Background(), query)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve data"})
		return
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		var word Word
		err = rows.Scan(&word.Word, &word.Definition, &word.Book, &word.Author, &word.Language, &word.Created_at)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to parse data"})
			return
		}
		words = append(words, word)
	}
	c.JSON(200, words)
}

func word_id(c *gin.Context) {
	id := c.Query("id")
	query := fmt.Sprintf("SELECT word, definition, book, author, language, created_at FROM words WHERE id = %s", id)
	rows, err := dbconn.Query(context.Background(), query)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve data"})
		return
	}
	defer rows.Close()

	var word Word
	for rows.Next() {
		err = rows.Scan(&word.Word, &word.Definition, &word.Book, &word.Author, &word.Language, &word.Created_at)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to parse data"})
			return
		}
	}
	c.JSON(200, word)
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
