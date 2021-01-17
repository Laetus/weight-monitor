package util

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

func Hello() string {
	str := "Hello World"
	log.Println(str)
	return str
}

type Config struct {
	OauthClientId string `split_words:"true"`
	EntriesLimit  int    `default:"100",split_words:true`
}

var ConfigInstance Config

func init() {
	err := envconfig.Process("", &ConfigInstance)
	if err != nil {
		log.Println("Config could not be loaded.")
		log.Fatal(err.Error())
	}
	log.Println("Config initialized")
}
