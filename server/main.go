package main

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	Text string `json:"text"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Pour dev local
	w.Header().Set("Content-Type", "application/json")

	msg := Message{Text: "Bonjour depuis le serveur Go"}
	json.NewEncoder(w).Encode(msg) // envoie JSON
}

func main() {
	http.HandleFunc("/api/hello", helloHandler)
	http.ListenAndServe(":8080", nil)
}
