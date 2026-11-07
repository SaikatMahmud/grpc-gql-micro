package main

import (
	"learn-grpc-gql-micro/account"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	var cgf Config
	err := envconfig.Process("", &cgf)
	if err != nil {
		log.Fatal(err)
	}

	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = account.NewPostgresRepository(cgf.DatabaseURL)

		if err != nil {
			log.Println(err)
		}
		return
	})

	defer r.Close()
	log.Println("Listening on :8080")

	s := account.NewService(r)
	log.Fatal(account.ListernGRPC(s, 8080))
}
