package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "Person_info"
)

type Person struct {
	ID int
	Name string
	Sex string
}


type DB struct {
	*sql.DB
}

func ConnectDB()(DB){
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+ "password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Invalid Database config", err)
	}

	if err = db.Ping(); err!=nil{
		log.Fatal("Databse not reachable",err)
	}
	fmt.Println("Successfully connected!")
	return DB{db}
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)

}

func DeleteUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	value, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid User Id" )
		return
	}

	DB := ConnectDB()

	_, err = DB.Exec("DELETE FROM person_data WHERE id = $1", value)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	DB.Close()

	respondJSON(w, http.StatusOK, map[string]string{"result":"success"})
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	value, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid User Id" )
		return
	}

	DB:= ConnectDB()
	row := DB.QueryRow("SELECT * FROM person_data WHERE id = $1", value)

	var person Person

	err = row.Scan(&person.ID, &person.Name, &person.Sex)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	DB.Close()

	respondJSON(w, http.StatusOK, person)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	DB:= ConnectDB()

	rows, err := DB.Query("SELECT * FROM person_data")
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	users := make([]Person, 0)

	for rows.Next() {
		var id int
		var name string
		var sex string
		user := Person{}
		err := rows.Scan(&id, &name, &sex)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		user.ID = id
		user.Name = name
		user.Sex = sex
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	DB.Close()

	respondJSON(w, http.StatusOK, users)
}


func CreateUser(w http.ResponseWriter, r *http.Request) {

	var person Person
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&person); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	DB := ConnectDB()

	err := DB.QueryRow("INSERT INTO person_data(name, sex) VALUES ($1, $2) RETURNING id", person.Name, person.Sex).Scan(&person.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	DB.Close()
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	value, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid User Id" )
		return
	}

	var person Person
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&person); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	DB := ConnectDB()

	_, err = DB.Exec("UPDATE person_data SET name = $1, sex = $2 WHERE id = $3", person.Name, person.Sex, value)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	DB.Close()

	respondJSON(w, http.StatusOK, map[string]string{"result":"success"})
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/users",GetAllUsers).Methods("GET");
	myRouter.HandleFunc("/user/{id}",GetUser).Methods("GET");
	myRouter.HandleFunc("/user",CreateUser).Methods("POST");
	myRouter.HandleFunc("/user/{id}",UpdateUser).Methods("PUT");
	myRouter.HandleFunc("/user/{id}",DeleteUsers).Methods("DELETE");
	log.Fatal(http.ListenAndServe(":9898", myRouter))
}

func main() {
	handleRequests()
}
