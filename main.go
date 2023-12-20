package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Task struct {
	ID    int    `json:"id"`
	Title  string `json:"title"`
	Description string `json:"description"`
}

func main() {
	//connect to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//create the table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS tasks (id SERIAL PRIMARY KEY, title TEXT, description TEXT)")

	if err != nil {
		log.Fatal(err)
	}

	//create router
	router := mux.NewRouter()
	router.HandleFunc("/api/tasks", getTasks(db)).Methods("GET")
	router.HandleFunc("/api/tasks/{id}", getTask(db)).Methods("GET")
	router.HandleFunc("/api/tasks", createTask(db)).Methods("POST")
	router.HandleFunc("/api/tasks/{id}", updateTask(db)).Methods("PUT")
	router.HandleFunc("/api/tasks/{id}", deleteTask(db)).Methods("DELETE")

	//start server
	log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router)))
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// get all users
func getTasks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM tasks")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		tasks := []Task{}
		for rows.Next() {
			var t Task
			if err := rows.Scan(&t.ID, &t.Title, &t.Description); err != nil {
				log.Fatal(err)
			}
			tasks = append(tasks, t)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(tasks)
	}
}

// get user by id
func getTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var t Task
		err := db.QueryRow("SELECT * FROM tasks WHERE id = $1", id).Scan(&t.ID, &t.Title, &t.Description)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(t)
	}
}

// create user
func createTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Task
		json.NewDecoder(r.Body).Decode(&t)

		err := db.QueryRow("INSERT INTO tasks (title, description) VALUES ($1, $2) RETURNING id", t.Title, t.Description).Scan(&t.ID)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(t)
	}
}

// update user
func updateTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Task
		json.NewDecoder(r.Body).Decode(&t)

		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE tasks SET title = $1, description = $2 WHERE id = $3", t.Title, t.Description, id)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(t)
	}
}

// delete user
func deleteTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var t Task
		err := db.QueryRow("SELECT * FROM tasks WHERE id = $1", id).Scan(&t.ID, &t.Title, &t.Description)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM tasks WHERE id = $1", id)
			if err != nil {
				//todo : fix error handling
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("Task deleted")
		}
	}
}
