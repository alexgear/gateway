package main

import (
	"log"
	"time"

	"github.com/alexgear/gateway/config"
	"github.com/alexgear/gateway/gservices"
)

var err error

func main() {
	var t0 time.Time
	var t1 time.Time

	cfg, err := config.New("config.toml")
	if err != nil {
		log.Fatalf(err.Error())
	}

	client := gservices.New()
	t0 = time.Now()
	log.Println(cfg.CalendarId)
	duty, err := gservices.GetDuty(client, cfg.CalendarId)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("%s is on duty today.\n", duty)
	t1 = time.Now()
	log.Printf("Spent %s.\n", t1.Sub(t0))

	// Spawn a bunch of goroutines to read messages
	c := make(chan gservices.Mail)
	for i := 0; i < 4; i++ {
		go func(c <-chan gservices.Mail) {
			for {
				email := <-c
				log.Println("Received", email)
				err = gservices.ReadMail(client, email)
				if err != nil {
					log.Println(err)
				}
			}
		}(c)
	}
	func(c chan<- gservices.Mail) {
		for {
			t0 = time.Now()
			mail, err := gservices.GetMail(client)
			if err != nil {
				log.Println(err)
			}
			for _, email := range mail {
				log.Println("Sending", email)
				c <- email
			}
			t1 = time.Now()
			log.Printf("Spent %s.\n", t1.Sub(t0))
			time.Sleep(1 * time.Second)
		}
	}(c)
	for {
	}
}
