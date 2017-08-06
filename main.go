package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
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
	go http.ListenAndServe(":2000", nil)

	r := Reader{[]string{"/home/colin/signal-cli/signal-cli-0.5.6/bin/signal-cli", "-u", "+12065391615"}, make(chan Message), sync.Mutex{}}
	go r.read()
	for v := range r.incoming {
		log.Printf("%v %q", v, v.Body)
		v.Destination = v.Source
		r.SendMessage(v)
	}
}
