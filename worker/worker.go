package worker

import (
	"log"
	"time"

	"github.com/alexgear/gateway/gservices"
)

var err error

// Spawn a bunch of goroutines to read messages
func InitWorker() {
	c := make(chan gservices.Mail)
	for i := 0; i < 4; i++ {
		go consumer(c)
	}
	go producer(c)
}

func producer(c chan<- gservices.Mail) {
	for {
		t0 := time.Now()
		mail, err := gservices.GetMail()
		if err != nil {
			log.Println(err)
		}
		for _, email := range mail {
			log.Println("Sending", email)
			c <- email
		}
		t1 := time.Now()
		log.Printf("Spent %s.\n", t1.Sub(t0))
		time.Sleep(10 * time.Second)
	}
}

func consumer(c <-chan gservices.Mail) {
	for {
		email := <-c
		log.Println("Received", email)
		err = gservices.ReadMail(email)
		if err != nil {
			log.Println(err)
		}
	}
}
