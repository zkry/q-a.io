package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Question represents a question along with it's vote count
type Question struct {
	ID   int    `json:"id"`
	Q    string `json:"q"`
	Vote int    `json:"vote"`
}

// Contains internal information that keeps track of users
type userKey string
type userInfo struct {
	room      string      // The room that the user registered with
	voteData  map[int]int // voteData keeps track of various votes user made
	questions []int       // List of questiosn asked by user
	email     string      // Keep track of user email
	isOwner   bool
}

type roomInfo struct {
	questions map[int]*Question
	isClosed  bool
}

var roomNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9]+([a-zA-Z0-9](_|-)?[a-zA-Z0-9])*[a-zA-Z0-9]*$`)

var users map[userKey]*userInfo
var rooms map[string]*roomInfo
var roomMux sync.Mutex
var questionID int = 0

func main() {
	entry := "./public/dist/index.html"
	static := "./public/dist/static"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	rooms = make(map[string]*roomInfo)
	users = make(map[userKey]*userInfo)

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1/").Subrouter()
	api.HandleFunc("/createRoom", createRoomHandler).Methods("POST")
	api.HandleFunc("/listRooms", createRoomHandler).Methods("POST")
	api.HandleFunc("/closeRoom/{roomName}", closeRoomHandler).Methods("POST")
	api.HandleFunc("/register/{roomName}", registerUser).Methods("POST")
	api.HandleFunc("/publishQuestion/{roomName}", publishQuestionHandler).Methods("POST")
	api.HandleFunc("/getQuestions/{roomName}", getQuestionsHandler).Methods("GET")
	api.HandleFunc("/vote/{roomName}", voteHandler).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(static))))

	r.PathPrefix("/").HandlerFunc(IndexHandler(entry))

	srv := &http.Server{
		Handler: handlers.LoggingHandler(os.Stdout, r),
		Addr:    ":" + port,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Starting server on port", port)
	log.Fatal(srv.ListenAndServe())
}

func IndexHandler(entrypoint string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("index: ", entrypoint)
		http.ServeFile(w, r, entrypoint)
	}
	return http.HandlerFunc(fn)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Accessing Home Page")
	w.WriteHeader(http.StatusOK)
}

func publishQuestionHandler(w http.ResponseWriter, r *http.Request) {
	const maxQLen = 5

	r.Header.Set("Content-Type", "application/json")
	roomName := strings.ToLower(mux.Vars(r)["roomName"])
	roomMux.Lock()
	defer roomMux.Unlock()

	if _, ok := rooms[roomName]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Room does not exist"}`))
		return
	}

	uID := r.FormValue("uID")
	if ui, ok := users[userKey(uID)]; !ok || ui.room != roomName {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"Must be registered to room"}`))
		return
	}

	if rooms[roomName].isClosed {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"status":"The room has been closed"}`))
		return
	}

	question := r.FormValue("question")
	if len(question) < maxQLen {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Question too short to be valad"}`))
	}

	questionID++
	rooms[roomName].questions[questionID] = &Question{ID: questionID, Q: question, Vote: 0}
	users[userKey(uID)].questions = append(users[userKey(uID)].questions, questionID)

	log.Printf("Add question \"%s\" to room %s\n", question, roomName)
	w.Write([]byte(`{"status":"ok"}`))
}

func voteHandler(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")

	roomName := strings.ToLower(mux.Vars(r)["roomName"])
	roomMux.Lock()
	defer roomMux.Unlock()

	if _, ok := rooms[roomName]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Room does not exist"}`))
		return
	}

	// Check if user is a valid, registered user
	uID := r.FormValue("uID")
	if ui, ok := users[userKey(uID)]; !ok || ui.room != roomName {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"Must be registered to room"}`))
		return
	}

	qID, err := strconv.Atoi(r.FormValue("qID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Could not read quesiton ID"}`))
		return
	}
	if _, ok := rooms[roomName].questions[qID]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"No question found with such ID"}`))
		return
	}

	qVal, err := strconv.Atoi(r.FormValue("val"))
	if err != nil || qVal < -1 || qVal > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Vote must be either 1, 0, or -1"}`))
		return
	}

	// Calculate the vote change needed for user selection
	prevVal := users[userKey(uID)].voteData[qID]
	diff := qVal - prevVal
	rooms[roomName].questions[qID].Vote += diff

	users[userKey(uID)].voteData[qID] = qVal

	log.Printf("Voted question %s/%d by %d", roomName, qID, qVal)
	w.Write([]byte(fmt.Sprintf(`{"status":"ok", "newCt": %d}`, rooms[roomName].questions[qID].Vote)))
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	roomName := strings.ToLower(mux.Vars(r)["roomName"])
	roomMux.Lock()
	defer roomMux.Unlock()
	if _, ok := rooms[roomName]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Room does not exist"}`))
		return
	}

	// Check the id passed in. If valid let the user keep it
	id := userKey(r.FormValue("uID"))
	if ui, ok := users[userKey(id)]; !ok || ui.room != roomName {
		id = generateUserID()
		users[id] = &userInfo{room: roomName, voteData: make(map[int]int)}
	}

	log.Printf("Registering user uuid '%s' with room %s\n", id, strings.ToLower(roomName))
	w.Write([]byte(fmt.Sprintf(`{"status":"ok","id":"%s"}`, id)))
}

func getQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	type QuestionList struct {
		Questions []Question `json:"questions"`
		IsClosed  bool       `json:"is_closed"`
	}
	r.Header.Set("Content-Type", "application/json")
	roomName := strings.ToLower(mux.Vars(r)["roomName"])
	roomMux.Lock()
	defer roomMux.Unlock()

	if _, ok := rooms[roomName]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Room does not exist"}`))
		return
	}

	resp := QuestionList{}
	resp.Questions = []Question{}
	for _, v := range rooms[roomName].questions {
		resp.Questions = append(resp.Questions, *v)
	}
	resp.IsClosed = rooms[roomName].isClosed

	b, err := json.Marshal(resp)
	if err != nil {
		log.Println("could not generate JSON from list of questions")
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(b)
}

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	roomName := r.FormValue("roomName")
	if !roomNameRegexp.MatchString(roomName) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Bad room name"}`))
		return
	}

	roomName = strings.ToLower(roomName)

	roomMux.Lock()
	defer roomMux.Unlock()

	// Check if room already exists
	if _, ok := rooms[roomName]; ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Room already exists"}`))
		return
	}

	rooms[roomName] = &roomInfo{questions: make(map[int]*Question)}
	secretCookie := http.Cookie{
		Name:   "questionKey",
		Value:  "123",
		MaxAge: 60 * 60 * 8,
	}
	http.SetCookie(w, &secretCookie)
	log.Println("Created room ", roomName)
	userID := generateUserID()
	users[userID] = &userInfo{room: roomName, voteData: make(map[int]int), isOwner: true}
	w.Write([]byte(fmt.Sprintf(`{"status":"ok", "uID":"%s"}`, string(userID))))
}

func closeRoomHandler(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")

	roomMux.Lock()
	defer roomMux.Unlock()

	roomName := strings.ToLower(mux.Vars(r)["roomName"])
	if _, ok := rooms[roomName]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Room does not exist"}`))
		return
	}

	uID := userKey(r.FormValue("uID"))
	if ui, ok := users[uID]; !ok || ui.room != roomName {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"Must be authorized to room"}`))
		return
	}

	// Check if user is the owner of the room
	if !users[uID].isOwner {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"Must be the owner of room to close it"}`))
		return
	}

	log.Println("Closing room ", roomName)
	rooms[roomName].isClosed = true
}

func listRoomsHandler(w http.ResponseWriter, r *http.Request) {
	type ListRoomsResp struct {
		Rooms []string `json:"rooms"`
	}

	w.Header().Set("Content-Type", "application/json")
	resp := ListRoomsResp{Rooms: make([]string, 0, len(rooms))}

	roomMux.Lock()
	defer roomMux.Unlock()

	for k := range rooms {
		resp.Rooms = append(resp.Rooms, k)
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Println("could not generate JSON from list of rooms")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func generateUserID() userKey {
	return userKey(uuid.New().String())
}
