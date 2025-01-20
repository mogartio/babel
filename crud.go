package test

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

const (
	ADD    = 1
	REMOVE = 2
	EDIT   = 3
	PRINT  = 4
	QUIT   = 5
)

func main() {
	connected, conn := connect()
	defer conn.Close(context.Background())
	if connected {
		for true {
			var action int
			fmt.Println("1: Add Word; 2: Remove Word; 3: Edit definition, 4: Print table, 5: quit ")
			fmt.Scan(&action)
			switch action {
			case ADD:
				insert_row(conn)
			case REMOVE:
				delete_word(conn)
			case EDIT:
				update_definition(conn)
			case PRINT:
				print_table(conn)
			case QUIT:
				return
			}
		}
	}
}

func delete_word(conn *pgx.Conn) {
	var word string
	fmt.Println("Type word to delete")
	fmt.Scan(&word)
	_, err := conn.Exec(context.Background(), "DELETE FROM words WHERE word = ($1)", word)
	if err != nil {
		log.Fatalf("Failed to delete word: %v", err)
	}
}

func insert_row(conn *pgx.Conn) {
	reader := bufio.NewScanner(os.Stdin)
	fmt.Println("Type a word")
	reader.Scan()
	word := reader.Text()
	fmt.Println("Type the definition of ", word)
	reader.Scan()
	definition := reader.Text()

	_, err := conn.Exec(context.Background(), "INSERT INTO words (word, definition) VALUES ($1, $2)", word, definition)
	if err != nil {
		log.Fatalf("Failed to add row: %v", err)
	}
}

// receives a word and its data. It adds a row using said data. On error it returns -1
func add_word(word string, definition string, book string, author string, language string) int {
	connected, conn := connect()
	defer conn.Close(context.Background())
	if !connected {
		return -1
	}
	_, err := conn.Exec(context.Background(),
		"INSERT INTO words (word, definition, book, author, language) VALUES ($1, $2, $3, $4, $5)", word, definition, book, author, language)
	if err != nil {
		log.Fatalf("Failed to add row: %v", err)
		return -1
	}
	return 1
}

func update_definition(conn *pgx.Conn) {
	var word string
	fmt.Println("Type word you want to update")
	fmt.Scan(&word)
	fmt.Println("Type new definition")
	reader := bufio.NewScanner(os.Stdin)
	reader.Scan()
	definition := reader.Text()
	_, err := conn.Exec(context.Background(), "UPDATE words SET definition = ($1) WHERE word = ($2)", definition, word)
	if err != nil {
		log.Fatalf("Failed to update: %v", err)
	}
}

func connect() (bool, *pgx.Conn) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return false, conn
	}

	// this tests the connection
	var version string
	if err := conn.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		log.Fatalf("Query failed: %v", err)
		return false, conn
	}

	log.Println("Connected to:", version)
	return true, conn
}

func print_table(conn *pgx.Conn) {
	rows, err := conn.Query(context.Background(), "SELECT id, word, definition FROM words")
	if err != nil {
		log.Fatalf("Quer failed: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var word string
		var definition string

		err := rows.Scan(&id, &word, &definition)
		if err != nil {
			log.Fatalf("Row scan failed: %v", err)
		}

		log.Printf("ID: %d, Word: %s, Meaning: %s\n", id, word, definition)
	}
}
