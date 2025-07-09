# Inquiry

Inquiry is a Go package that converts CSV files into a SQLite database, allowing you to run SQL statements on them.

## Table of contents

1. [How to install](#how-to-install)
2. [How to use](#how-to-use)
    1. [Defining your struct](#defining-your-struct)
        1. [Go field to SQLite column mapping](#go-field-to-sqlite-column-mapping)
    2. [Creating an in-memory SQLite database from a CSV file](#creating-an-in-memory-sqlite-database-from-a-csv-file)
        1. [Without options](#without-options)
        2. [With options](#with-options)
    3. [Creating a new table from a CSV file and adding it to an existing SQLite database](#creating-a-new-table-from-a-csv-file-and-adding-it-to-an-existing-sqlite-database)
        1. [Adding a table to an in-memory database from a CSV](#adding-a-table-to-an-in-memory-database-from-a-csv)
        2. [Adding a table to an on-disk database from a CSV](#adding-a-table-to-an-on-disk-database-from-a-csv)
3. [Contributing](#contributing)

## How to install

`go get github.com/sionpixley/inquiry`

## How to use

Using Inquiry is pretty simple: You "connect" to the CSV file and Inquiry will return a `*sql.DB` and an `error`. You can then use the returned `*sql.DB` to do any operations that you would normally do with a SQLite database.

You can also create new tables from CSV files and add them to an existing SQLite database (in-memory or not).

### Defining your struct

Inquiry uses generics and package `reflect` to build the SQLite database/table and to insert the data. Please put the struct fields in the same position as the column they are supposed to represent in the CSV file. Different struct definitions yield different SQL `CREATE TABLE` statements. For example:

```go
type Student struct {
    Id         int
    FirstName  string
    MiddleName *string
    LastName   string
    IsFullTime bool
    GPA        float64
}
```

```sql
-- SQL CREATE TABLE statement that is generated from the above Go struct definition.
CREATE TABLE 'Student'(
    'Id'         INTEGER NOT NULL,
    'FirstName'  TEXT NOT NULL,
    'MiddleName' TEXT NULL,
    'LastName'   TEXT NOT NULL,
    'IsFullTime' INTEGER NOT NULL CHECK('IsFullTime' IN (0,1)),
    'GPA'        REAL NOT NULL
);
```

Please consult the table below for a full list of which Go field types map to which SQLite column types.

#### Go field to SQLite column mapping

| Go Field Type | SQLite Column Type |
| ------------- | --------------- |
| `bool` | `INTEGER NOT NULL CHECK(<field_name> IN (0,1))` |
| `*bool` | `INTEGER NULL CHECK(<field_name> IN (0,1))` |
| - `float32` <br> - `float64` | `REAL NOT NULL` |
| - `*float32` <br> - `*float64` | `REAL NULL` |
| - `int` <br> - `int8` <br> - `int16` <br> - `int32` <br> - `int64` | `INTEGER NOT NULL` |
| - `*int` <br> - `*int8` <br> - `*int16` <br> - `*int32` <br> - `*int64` | `INTEGER NULL` |
| `string` | `TEXT NOT NULL` |
| `*string` | `TEXT NULL` |

On nullable columns, certain values in the CSV will insert a `NULL` into the database. These values are: an empty value, `null`, and `NULL`. On non-nullable columns, these values can potentially throw an error or be inserted as they are (especially in the case of a `TEXT` column).

#### Inquiry struct tags

You can use struct tags to create primary keys, indexes, and unique constraints.

> **Note:** Inquiry currently only supports single-column indexes and unique constraints.

```go
type Customer struct {
    Index            int    `inquiry:"primaryKey"`
    CustomerId       string `inquiry:"unique"`
    FirstName        string
    LastName         string
    Company          string
    City             string
    Country          string
    Phone1           string
    Phone2           string
    Email            *string
    SubscriptionDate string
    Website          string `inquiry:"index"`
}
```

```sql
-- SQL that is generated from the above Go struct definition.
CREATE TABLE 'Customer'(
    'Index'            INTEGER NOT NULL,
    'CustomerId'       TEXT NOT NULL,
    'FirstName'        TEXT NOT NULL,
    'LastName'         TEXT NOT NULL,
    'Company'          TEXT NOT NULL,
    'City'             TEXT NOT NULL,
    'Country'          TEXT NOT NULL,
    'Phone1'           TEXT NOT NULL,
    'Phone2'           TEXT NOT NULL,
    'Email'            TEXT NULL,
    'SubscriptionDate' TEXT NOT NULL,
    'Website'          TEXT NOT NULL,

    CONSTRAINT PK_Customer_Index PRIMARY KEY('Index'),
    CONSTRAINT Unique_Customer_CustomerId UNIQUE('CustomerId')
);

CREATE INDEX NonClustered_Customer_Website ON 'Customer'('Website');
```

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

```go
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
    db, err := inquiry.Connect[Example]("example.csv")
    if err != nil {
        log.Fatalln(err.Error())
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

```go
type CsvOptions struct {
    CommentCharacter rune `json:"commentCharacter"`
    Delimiter        rune `json:"delimiter"`
    HasHeaderRow     bool `json:"hasHeaderRow"`
    TrimLeadingSpace bool `json:"trimLeadingSpace"`
    UseLazyQuotes    bool `json:"useLazyQuotes"`
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

```go
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
    options := inquiry.CsvOptions{
        Delimiter:    '|',
        HasHeaderRow: true,
    }
    
    db, err := inquiry.ConnectWithOptions[Example]("example.csv", options)
    if err != nil {
        log.Fatalln(err.Error())
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

```go
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
    
    db, err := inquiry.ConnectWithOptions[Example]("example.csv", options)
    if err != nil {
        log.Fatalln(err.Error())
    }
    // Don't forget to close the database.
    defer db.Close()
    
    err = inquiry.CreateTable[Test](db, "test.csv")
    if err != nil {
        log.Fatalln(err.Error())
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

```go
// main.go

package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/mattn/go-sqlite3"
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
    
    err = inquiry.CreateTableWithOptions[Test](db, "test.csv", options)
    if err != nil {
        log.Fatalln(err.Error())
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

## Contributing

All contributions are welcome! If you wish to contribute to the project, the best way would be forking this repo and making a pull request from your fork with all of your suggested changes.
