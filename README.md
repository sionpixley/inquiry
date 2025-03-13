# Inquiry

Inquiry is a Go package that converts CSV files into a SQLite database, allowing you to run SQL statements on them.

> **Note:** Inquiry is *very* new. It's still just a pet project. It has not been "battle-tested". It has not been benchmarked. It has not been stress tested. The API is still unstable and breaking changes can happen at any time. I would not recommend building any production-ready code with it. Use at your own risk.

## Table of contents

1. [Project structure](#project-structure)
2. [How to install](#how-to-install)
3. [How to use](#how-to-use)
    1. [Creating an in-memory SQLite database from a CSV file](#creating-an-in-memory-sqlite-database-from-a-csv-file)
        1. [Without options](#without-options)
        2. [With options](#with-options)
    2. [Creating a new table from a CSV file and adding it to an existing SQLite database](#creating-a-new-table-from-a-csv-file-and-adding-it-to-an-existing-sqlite-database)
        1. [Adding a table to an in-memory database from a CSV](#adding-a-table-to-an-in-memory-database-from-a-csv)
        2. [Adding a table to an on-disk database from a CSV](#adding-a-table-to-an-on-disk-database-from-a-csv)

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

You can also create new tables from CSV files and add them to an existing SQLite database (in-memory or not). 

### Creating an in-memory SQLite database from a CSV file

To create an in-memory database from a CSV file, use the `Connect` or `ConnectWithOptions` function. 

With options, you can specify your CSV delimiter and whether the file has a header row or not. If you don't provide options, Inquiry will default the delimiter to a comma and assume there is no header row.

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
    db, errs := inquiry.Connect[Example]("example.csv")
    if errs != nil {
        for _, e := range errs {
            log.Println(e.Error())
        }
        os.Exit(1)
    }
    // Don't forget to close the database.
    defer db.Close()

    rows, err := db.Query("SELECT * FROM Example WHERE Value > 80 ORDER BY Name ASC;")
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

```
// output

4 happy 100.500000
3 yo 90.300000
```

#### With options

The options are set using a struct: `CsvOptions`. Please see below for the definition of the `CsvOptions` struct and an example of using it.

```
type CsvOptions struct {
    Delimiter    rune `json:"delimiter"`
    HasHeaderRow bool `json:"hasHeaderRow"`
}
```

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
    options := inquiry.CsvOptions{
        Delimiter:    '|',
        HasHeaderRow: true,
    }

    // 'errs' is a []error with a cap of 25. If there are no errors, then 'errs' will be nil.
    db, errs := inquiry.ConnectWithOptions[Example]("example.csv", options)
    if errs != nil {
        for _, e := range errs {
            log.Println(e.Error())
        }
        os.Exit(1)
    }
    // Don't forget to close the database.
    defer db.Close()

    rows, err := db.Query("SELECT * FROM Example WHERE Value > 80 ORDER BY Name ASC;")
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

```
// output

4 happy 100.500000
3 yo 90.300000
```

### Creating a new table from a CSV file and adding it to an existing SQLite database

To create a new table from a CSV file and add it to an existing SQLite database, use the `CreateTable` or `CreateTableWithOptions` function.

With options, you can specify your CSV delimiter and whether the file has a header row or not. If you don't provide options, Inquiry will default the delimiter to a comma and assume there is no header row.

This works on in-memory databases as well as databases that persist to disk.

#### Adding a table to an in-memory database from a CSV

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
// test.csv

1,this is a horrible test
2,ehhh
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

type Test struct {
    Id  int
    Val string
}

func main() {
    options := inquiry.CsvOptions{
        Delimiter:    '|',
        HasHeaderRow: true,
    }

    // 'errs' is a []error with a cap of 25. If there are no errors, then 'errs' will be nil.
    db, errs := inquiry.ConnectWithOptions[Example]("example.csv", options)
    if errs != nil {
        for _, e := range errs {
            log.Println(e.Error())
        }
        os.Exit(1)
    }
    // Don't forget to close the database.
    defer db.Close()

    // 'errs' is a []error with a cap of 25. If there are no errors, then 'errs' will be nil.
    errs = inquiry.CreateTable[Test](db, "test.csv")
    if errs != nil {
        for _, e := range errs {
            log.Println(e.Error())
        }
        os.Exit(1)
    }

    rows, err := db.Query("SELECT * FROM Example WHERE Value > 80 ORDER BY Name ASC;")
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

    rows, err = db.Query("SELECT * FROM Test ORDER BY Id DESC;")
    if err != nil {
        log.Fatalln(err.Error())
    }

    for rows.Next() {
        var test Test
        err = rows.Scan(&test.Id, &test.Val)
        if err != nil {
            log.Fatalln(err.Error())
        }
        fmt.Printf("%d %s", test.Id, test.Val)
        fmt.Println()
    }
}
```

```
// output

4 happy 100.500000
3 yo 90.300000
2 ehhh
1 this is a horrible test
```

#### Adding a table to an on-disk database from a CSV

```
// example.db is an on-disk database with data equivalent to:

1,hi,2.8
2,hello,3.4
3,yo,90.3
4,happy,100.5
5,yay,8.1
```

```
// test.csv

1;this is a horrible test
2;ehhh
```

```
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/ncruces/go-sqlite3/driver"
    _ "github.com/ncruces/go-sqlite3/embed"
    "github.com/sionpixley/inquiry/pkg/inquiry"
)

type Example struct {
    Id    int
    Name  string
    Value float64
}

type Test struct {
    Id  int
    Val string
}

func main() {
    db, err := sql.Open("sqlite3", "example.db")
    if err != nil {
        log.Fatalln(err.Error())
    }
    // Don't forget to close the database.
    defer db.Close()

    options := inquiry.CsvOptions{
        Delimiter:    ';',
        HasHeaderRow: false,
    }

    // 'errs' is a []error with a cap of 25. If there are no errors, then 'errs' will be nil.
    errs := inquiry.CreateTableWithOptions[Test](db, "test.csv", options)
    if errs != nil {
        for _, e := range errs {
            log.Println(e.Error())
        }
        os.Exit(1)
    }

    rows, err := db.Query("SELECT * FROM Example WHERE Value > 80 ORDER BY Name ASC;")
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

    rows, err = db.Query("SELECT * FROM Test ORDER BY Id DESC;")
    if err != nil {
        log.Fatalln(err.Error())
    }

    for rows.Next() {
        var test Test
        err = rows.Scan(&test.Id, &test.Val)
        if err != nil {
            log.Fatalln(err.Error())
        }
        fmt.Printf("%d %s", test.Id, test.Val)
        fmt.Println()
    }
}
```

```
// output

4 happy 100.500000
3 yo 90.300000
2 ehhh
1 this is a horrible test
```
