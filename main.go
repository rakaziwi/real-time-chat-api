package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/gorilla/websocket"
)

// Define our message object
type Message struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var messages []Message // array to store sent message

func main() {
	// Start listening for incoming chat messages
	go handleMessages()

	setRoutes()
}

// Set the application routes
func setRoutes() {
	// Set routes
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/ws", handleConnections)
	router.HandleFunc("/api/messages", getMessages).Methods("GET")
	router.HandleFunc("/api/messages", sendMessage).Methods("POST")

	// Handling the static page in SPA
	spa := spaHandler{staticPath: "public", indexPath: "index.html"} // Set default page folder
	router.PathPrefix("/").Handler(spa)

	log.Println("http server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// ======================================== API

// Status response object for generalize api response
type StatusResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Message response object
type MessageResponse struct {
	Status StatusResponse `json:"status"`
	Data   []Message      `json:"data"`
}

// Get random username for send message api
func getRandUsername() string {
	usernames := [9]string{"Unyil", "Cuplis", "Pak Ogah", "Cloud", "Tifa", "Aerith", "Kiriyama", "Hinata", "Akari"}
	chosenone := usernames[rand.Intn(len(usernames))]

	return chosenone
}

// Utility to generalize api response
func getMessageResponse(w http.ResponseWriter, messages []Message, err error) {
	w.Header().Set("Content-Type", "application/json")

	var statusResponse StatusResponse
	var messageResponse MessageResponse
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		statusResponse.Error = true
		statusResponse.Message = err.Error()
		statusResponse.Code = 400
	} else {
		w.WriteHeader(http.StatusOK)
		statusResponse.Error = false
		statusResponse.Message = ""
		statusResponse.Code = 200
	}

	messageResponse.Status = statusResponse
	messageResponse.Data = messages

	json.NewEncoder(w).Encode(messageResponse)
}

// Retrieve all sent message via api
func getMessages(w http.ResponseWriter, r *http.Request) {
	getMessageResponse(w, messages, nil)
}

// Send message via api
func sendMessage(w http.ResponseWriter, r *http.Request) {
	// Store the message
	var message Message
	_ = json.NewDecoder(r.Body).Decode(&message)
	message.Username = getRandUsername()
	storeMessage(message)

	// Broadcast the message
	broadcast <- message

	var msgs []Message
	msgs = append(msgs, message)
	getMessageResponse(w, msgs, nil)
}

// Store sent message
func storeMessage(message Message) {
	message.ID = strconv.Itoa(rand.Intn(1000000))
	messages = append(messages, message)
}

// ======================================== Web Socket
var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// Configure the upgrader for request to ws
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Join the ws hub and handle income new send message
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// If the username not null, store the sent message
		if msg.Username != "" {
			storeMessage(msg)
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

// Handle incoming broadcasted message and send it to all connected client
func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast

		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// ======================================== SPA
type spaHandler struct {
	staticPath string
	indexPath  string
}

// SPA handler
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
