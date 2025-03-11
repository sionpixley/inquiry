# Inquiry

Inquiry is a Go package that converts a CSV file into an in-memory SQLite database, allowing you to run SQL statements on the CSV file.

> **Note:** Inquiry is *very* new. It's still just a pet project. It has not been "battle-tested". It has not been benchmarked. It has not been stress tested. The API is still unstable and breaking changes can happen at any time. I would not recommend building any production-ready code with it. Use at your own risk.

## Table of contents

1. [Project structure](#project-structure)
2. [How to install](#how-to-install)
3. [How to use](#how-to-use)
    1. [Examples](#examples)
        1. [Without options](#without-options)
        2. [With options](#with-options)

## Project structure

```
.
├── CODE_OF_CONDUCT.md
├── LICENSE
├── README.md
├── SECURITY.md
├── go.mod
├── go.sum
└── pkg
    └── inquiry
        ├── helpers.go
        └── inquiry.go
```

## How to install

`go get github.com/sionpixley/inquiry`

## How to use

Using Inquiry is pretty simple: You "connect" to the CSV file and Inquiry will return a `*sql.DB` and a `[]error`. You can then use the returned `*sql.DB` to do any operations that you would normally do with a SQLite database.

### Examples

You can connect to your CSV file with or without options (`ConnectWithOptions` or just `Connect`). With options, you can specify your CSV delimiter and whether the file has a header row or not. If you don't provide options, Inquiry will default the delimiter to a comma and assumes there is no header row.

#### Without options

```
// example.csv

1,hi,2.8
2,hello,3.4
3,yo,90.3
4,happy,100.5
5,yay,8.1
```

```
// main.go

package main

import (
    "fmt"
    "log"
    "os"

    "github.com/sionpixley/inquiry/pkg/inquiry"
)

type Example struct {
    Id    int
    Name  string
    Value float64
}

func main() {
    // 'errs' is a []error with a cap of 25. If there are no errors, then 'errs' will be nil.
    csvFile, errs := inquiry.Connect[Example]("example.csv")
    if errs != nil {
        for _, e := range errs {
            log.Println(e.Error())
        }
        os.Exit(1)
    }
    // Don't forget to close the database.
    defer csvFile.Close()

    rows, err := csvFile.Query("SELECT * FROM Example WHERE Value > 80 ORDER BY Name ASC;")
    if err != nil {
        log.Fatalln(err.Error())
    }

    for rows.Next() {
        var example Example
        err = rows.Scan(&example.Id, &example.Name, &example.Value)
        if err != nil {
            log.Fatalln(err.Error())
        }
        fmt.Printf("%d %s %f", example.Id, example.Name, example.Value)
        fmt.Println()
    }
}
```

#### With options

```
// example.csv

Id|Name|Value
1|hi|2.8
2|hello|3.4
3|yo|90.3
4|happy|100.5
5|yay|8.1
```

```
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/sionpixley/inquiry/pkg/inquiry"
)

type Example struct {
    Id    int
    Name  string
    Value float64
}

func main() {
    options := inquiry.InquiryOptions{
        Delimiter:    '|',
        HasHeaderRow: true,
    }

    // 'errs' is a []error with a cap of 25. If there are no errors, then 'errs' will be nil.
    csvFile, errs := inquiry.ConnectWithOptions[Example]("example.csv", options)
    if errs != nil {
        for _, e := range errs {
            log.Println(e.Error())
        }
        os.Exit(1)
    }
    // Don't forget to close the database.
    defer csvFile.Close()

    rows, err := csvFile.Query("SELECT * FROM Example WHERE Value > 80 ORDER BY Name ASC;")
    if err != nil {
        log.Fatalln(err.Error())
    }

    for rows.Next() {
        var example Example
        err = rows.Scan(&example.Id, &example.Name, &example.Value)
        if err != nil {
            log.Fatalln(err.Error())
        }
        fmt.Printf("%d %s %f", example.Id, example.Name, example.Value)
        fmt.Println()
    }
}
```
