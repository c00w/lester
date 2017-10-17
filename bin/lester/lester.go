package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/c00w/lester"
)

func main() {
	r := lester.NewKeybaseReader()

	b, err := lester.NewBoltMemory("/home/colin/boltmemory.db")
	if err != nil {
		log.Fatalf("Error opening bolt db: %v", err)
	}
	h := lester.NewHandler(r, lester.EchoBrain{W: r}, lester.FinanceBrain{W: r, M: b})
	defer h.Close()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	select {
	case s := <-c:
		log.Printf("Received %v", s)
		return
	}
}
