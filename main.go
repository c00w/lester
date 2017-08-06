package main

import (
	"fmt"
	"log"
	"net/http"
)

func call(w http.ResponseWriter, r *http.Request) {
	log.Print("Twilio call - redirecting")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?><Response><Dial>+1-206-930-0074</Dial></Response>`)
}

func twilio(w http.ResponseWriter, r *http.Request) {
	log.Print("Twilio")
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		return
	}
	log.Print(r.PostForm)
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?><Response></Response>`)
}

func handle(w http.ResponseWriter, r *http.Request) {
	log.Print(r)
	fmt.Fprintf(w, "Hello World")
}

func main() {
	http.HandleFunc("/call", call)
	http.HandleFunc("/sms", twilio)
	http.HandleFunc("/", handle)
	http.ListenAndServe(":2000", nil)
}
