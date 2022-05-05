package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const PersistentFile = "db/persistent-pies.json"

type PieContent struct {
	Members []Member `json:"members"`
	Pies    []Pie    `json:"pies"`
	State   State    `json:"state"`
}

type Member struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Pie struct {
	Type   string `json:"type"`
	Date   string `json:"date"`
	Member Member `json:"member"`
}

type State struct {
	Current  Order `json:"current"`
	Next     Order `json:"next"`
	Previous Order `json:"previous"`
}

type Order struct {
	Date string `json:"date"`
	ID   string `json:"id"`
}

func LoadPersistentData(file string) PieContent {
	jsonFile, err := os.Open(PersistentFile)
	if err != nil {
		jsonFile, err = os.Create(file)
		if err != nil {
			fmt.Println("Oh no..")
			return PieContent{}
		}
	}

	fmt.Println("Successfully Opened users.json")
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var pie_ctx PieContent

	json.Unmarshal(byteValue, &pie_ctx)

	return pie_ctx
}

func SavePersistentDate(pies *PieContent) {
	content, _ := json.Marshal(pies)
	err := ioutil.WriteFile(PersistentFile, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
