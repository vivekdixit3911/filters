package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Database connection function
func dbConn() (*sql.DB, error) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "A800900plmA"
	dbName := "filters"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(localhost:3306)/"+dbName)
	if err != nil {
		return nil, err
	}
	// Test the database connection
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// Enable CORS
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Format date to DD-MM-YYYY
func formatDate(date string) string {
	parsedDate, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		parsedDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			return date
		}
	}
	return parsedDate.Format("02-01-2006")
}

// API Endpoint for Fetching Filtered Data
func getData(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	// Connect to the database
	db, err := dbConn()
	if err != nil {
		http.Error(w, "Database connection error: "+err.Error(), http.StatusInternalServerError)
		log.Println("Database connection error:", err)
		return
	}
	defer db.Close()

	// Get date parameters
	fromDate := r.URL.Query().Get("fromDate")
	toDate := r.URL.Query().Get("toDate")

	// Log received values
	fmt.Println("Received fromDate:", fromDate, "toDate:", toDate)

	query := "SELECT * FROM customerdata"
	var args []interface{}

	// Apply filtering only if both dates are provided
	if fromDate != "" && toDate != "" {
		query += " WHERE DATE(created_at) BETWEEN ? AND ?"
		args = append(args, fromDate, toDate)
	}

	fmt.Println("Executing SQL Query:", query, args) // Debugging log

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		log.Println("SQL Query Error:", err)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Error fetching columns: "+err.Error(), http.StatusInternalServerError)
		log.Println("Error fetching columns:", err)
		return
	}

	var data []map[string]string
	for rows.Next() {
		values := make([]sql.RawBytes, len(columns))
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
			log.Println("Error scanning row:", err)
			return
		}

		row := make(map[string]string)
		for i, col := range values {
			if columns[i] == "created_at" || columns[i] == "updated_at" || columns[i] == "date_of_birth" {
				row[columns[i]] = formatDate(string(col))
			} else {
				row[columns[i]] = string(col)
			}
		}
		data = append(data, row)
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "JSON Encoding Error: "+err.Error(), http.StatusInternalServerError)
		log.Println("JSON Encoding Error:", err)
	}
}

func main() {
	http.HandleFunc("/fetch", getData)
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
