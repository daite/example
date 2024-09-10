package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Student struct {
	Id    int
	Name  string
	Age   int
	Score int
}

type Students []Student

var students map[int]Student
var lastId int

func (s Students) Len() int {
	return len(s)
}

func (s Students) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Students) Less(i, j int) bool {
	return s[i].Id < s[j].Id
}

func init() {
	students = make(map[int]Student)
	students[1] = Student{1, "aaa", 16, 87}
	students[2] = Student{2, "bbb", 18, 98}
	lastId = 2
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log incoming request
		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
		log.Printf("Headers: %v", r.Header)

		// Capture response body for logging
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK, body: &bytes.Buffer{}}

		// Start timer
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(rec, r)

		// Log outgoing response
		duration := time.Since(start)
		log.Printf("Response status: %d", rec.statusCode)
		log.Printf("Response body: %s", rec.body.String())
		log.Printf("Response time: %s\n", duration)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b) // capture the body for logging
	return r.ResponseWriter.Write(b)
}

func MakeWebHandler() http.Handler {
	mux := mux.NewRouter()
	mux.HandleFunc("/students", GetStudentListHandler).Methods("GET")
	mux.HandleFunc("/students/{id:[0-9]+}", GetStudentHandler).Methods("GET")
	mux.HandleFunc("/students", PostStudentHandler).Methods("POST")
	return loggingMiddleware(mux)
}

func GetStudentListHandler(w http.ResponseWriter, r *http.Request) {
	list := make(Students, 0)
	for _, student := range students {
		list = append(list, student)
	}
	sort.Sort(list)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	prettyJSON, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(prettyJSON)
}

func GetStudentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	student, ok := students[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	prettyJSON, err := json.MarshalIndent(student, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(prettyJSON)
}

func main() {
	fmt.Println("Server is running on port 3000...")
	http.ListenAndServe(":3000", MakeWebHandler())
}

func PostStudentHandler(w http.ResponseWriter, r *http.Request) {
	var student Student
	err := json.NewDecoder(r.Body).Decode(&student)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	lastId++
	student.Id = lastId
	students[lastId] = student
	w.WriteHeader(http.StatusCreated)
}
