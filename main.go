package main

import (
	"fmt"
	"log"

	"github.com/sionpixley/inquiry/pkg/inquiry"
)

type Example struct {
	Id   int
	Name string
	Test float64
}

func main() {
	err := inquiry.Connect[Example]("example.csv", false)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer inquiry.Close()

	row := inquiry.Database.QueryRow("SELECT * FROM Example;")
	var example Example
	err = row.Scan(&example.Id, &example.Name, &example.Test)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println(example.Id)
	fmt.Println(example.Name)
	fmt.Println(example.Test)
}
