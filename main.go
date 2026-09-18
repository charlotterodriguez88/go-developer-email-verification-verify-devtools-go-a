package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	client, err := NewInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var user Signup
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil || user.Email == "" || user.UserID == "" {
			http.Error(w, "email and user_id are required", http.StatusBadRequest)
			return
		}
		result, err := SendVerification(client, "https://developers.example.com", user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message_id": result.MessageID, "status": "verification_sent"})
	})
	log.Println("signup service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
