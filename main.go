package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
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
	room     string      // The room that the user registered with
	voteData map[int]int // voteData keeps track of various votes user made
	email    string      // Keep track of user email
}

var roomNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9]+([a-zA-Z0-9](_|-)?[a-zA-Z0-9])*[a-zA-Z0-9]*$`)

var users map[userKey]*userInfo
var rooms map[string]map[int]*Question
var roomMux sync.Mutex
var questionID int = 0

func main() {
	r := mux.NewRouter()

	rooms = make(map[string]map[int]*Question)
	users = make(map[userKey]*userInfo)

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/createRoom", createRoomHandler).Methods("POST")
	r.HandleFunc("/listRooms", listRoomsHandler).Methods("GET")
	r.HandleFunc("/register/{roomName}", registerUser).Methods("POST")
	r.HandleFunc("/publishQuestion/{roomName}", publishQuestionHandler).Methods("POST")
	r.HandleFunc("/getQuestions/{roomName}", getQuestionsHandler).Methods("GET")
	r.HandleFunc("/vote/{roomName}", voteHandler).Methods("GET")
	http.Handle("/", r)

	log.Println("Starting server on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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

	question := r.FormValue("question")
	if len(question) < maxQLen {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"Question too short to be valad"}`))
	}

	questionID++
	rooms[roomName][questionID] = &Question{ID: questionID, Q: question, Vote: 0}

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
	if _, ok := rooms[roomName][qID]; !ok {
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
	rooms[roomName][qID].Vote += diff

	users[userKey(uID)].voteData[qID] = qVal

	log.Printf("Voted question %s/%d by %d", roomName, qID, qVal)
	w.Write([]byte(fmt.Sprintf(`{"status":"ok", "newCt": %d}`, rooms[roomName][qID].Vote)))
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

	id := userKey(uuid.New().String())
	log.Printf("Registering user uuid '%s' with room %s\n", id, strings.ToLower(roomName))
	users[id] = &userInfo{room: roomName, voteData: make(map[int]int)}
	w.Write([]byte(fmt.Sprintf(`{"status":"ok","id":"%s"}`, id)))
}

func getQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	type QuestionList struct {
		Questions []Question `json:"questions"`
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
	for _, v := range rooms[roomName] {
		resp.Questions = append(resp.Questions, *v)
	}

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

	rooms[roomName] = make(map[int]*Question)
	secretCookie := http.Cookie{
		Name:   "questionKey",
		Value:  "123",
		MaxAge: 60 * 60 * 8,
	}
	http.SetCookie(w, &secretCookie)
	log.Println("Created room ", roomName)
	w.Write([]byte(`{"status":"ok"}`))
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
