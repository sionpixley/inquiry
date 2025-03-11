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
	csvFile, errs := inquiry.Connect[Example]("example.csv")
	if errs != nil {
		for _, e := range errs {
			log.Println(e.Error())
		}
		os.Exit(1)
	}
	defer csvFile.Close()

	rows, err := csvFile.Query("SELECT * FROM Example WHERE Test > 80 ORDER BY Name ASC;")
	if err != nil {
		log.Fatalln(err.Error())
	}

	for rows.Next() {
		var example Example
		err = rows.Scan(&example.Id, &example.Name, &example.Test)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("%d %s %f", example.Id, example.Name, example.Test)
		fmt.Println()
	}
}
