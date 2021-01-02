package main

// Ingress to the entire system
// - recieve requests
// - boot pod to serve requests if scaled to 0
// - spawns client somwhow somewhere
// - routing happens and http->grpc happens too

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Update to GRPC
func serveMessage(w http.ResponseWriter, r *http.Request) {
	msg := os.Getenv("API_MSG")

	log.Println(msg)
	fmt.Fprintf(w, "%+v", string(msg))
}

func main() {
	log.Println("Booting API Service")
	http.HandleFunc("/", serveMessage)
	log.Fatal(http.ListenAndServe(":80", nil))
}
