package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/fly-apps/litevfs-demo/fly/pkg/litevfs"
)

//go:embed html/*
var static embed.FS

func main() {
	db, err := sql.Open("litevfs", "demo.db")
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}

	if err := litevfs.WithWriteLease(db, migrate); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			Latency string `json:"latency"`
		}

		now := time.Now()
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := litevfs.WithWriteLease(db, func(db *sql.DB) error {
			_, err := db.Exec("INSERT INTO data (data) VALUES(?)", rand.Int())
			return err
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		latency := time.Since(now)

		json.NewEncoder(w).Encode(response{Latency: latency.String()})
	})

	http.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		type Record struct {
			ID    int `json:"id"`
			Value int `json:"value"`
		}

		type response struct {
			Latency string   `json:"latency"`
			Records []Record `json:"records"`
		}

		now := time.Now()
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query("SELECT * FROM (SELECT * FROM data ORDER BY id DESC LIMIT 20) ORDER BY id ASC")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var resp response
		for rows.Next() {
			var r Record
			if err := rows.Scan(&r.ID, &r.Value); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp.Records = append(resp.Records, r)

		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		latency := time.Since(now)
		resp.Latency = latency.String()

		json.NewEncoder(w).Encode(resp)
	})

	sub, err := fs.Sub(static, "html")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(sub)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on port %s", port)
	http.ListenAndServe(":"+port, nil)
}

func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("start a TX: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var version int
	if err := tx.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf("read user_version: %w", err)
	}
	nextVersion := version + 1

	for nextVersion < len(migrations) {
		if _, err := tx.Exec(migrations[nextVersion]); err != nil {
			return fmt.Errorf("run migration %d: %w", nextVersion, err)
		}
		nextVersion += 1
	}

	if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, nextVersion-1)); err != nil {
		return fmt.Errorf("set user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit TX: %w", err)
	}

	return nil
}

var migrations = []string{
	``,
	`CREATE TABLE data (id INTEGER PRIMARY KEY AUTOINCREMENT, data INTEGER)`,
}
