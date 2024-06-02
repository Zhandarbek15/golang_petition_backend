package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"petition_api/internal/app/apiserver"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "conf", "configs/config.json", "config path")
}

func main() {
	flag.Parse()
	// Чтение содержимого JSON-файла
	file, err := ioutil.ReadFile(configPath)
	errHandler(err)

	config := apiserver.NewConfig()

	// Распаковка JSON в структуру
	err = json.Unmarshal(file, config)
	errHandler(err)

	s := apiserver.NewApiServer(config)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}

func errHandler(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
