package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sionpixley/inquiry/pkg/inquiry"
)

type Example struct {
	Id   int
	Name string
	Test float64
}

func main() {
	errs := inquiry.Connect[Example]("example.csv")
	if errs != nil {
		for _, e := range errs {
			log.Println(e.Error())
		}
		os.Exit(1)
	}
	defer inquiry.Close()

	row := inquiry.Database.QueryRow("SELECT * FROM Example;")
	var example Example
	err := row.Scan(&example.Id, &example.Name, &example.Test)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println(example.Id)
	fmt.Println(example.Name)
	fmt.Println(example.Test)
}
