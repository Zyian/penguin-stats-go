# Penguin Stats API Wrapper
[![API reference](https://img.shields.io/badge/godoc-reference-5272B4)](https://pkg.go.dev/github.com/Zyian/penguin-stats-go?tab=doc)

An API wrapper for the Arknights item drop analytics website: https://penguin-stats.io

## Supported Functionality

- Reports
  - Submit Reports
  - Recall Reports


## TODO

- Matrix API parsing
- Better error handling

## Usage

Easiest way to isntall `penguin-stats-go` is through the `go get` command:
```term
$ go get -u github.com/Zyian/penguin-stats-go
```

You can then use it to submit your item drop reports programmatically:
```go
package main

import (
    penguin "github.com/Zyian/penguin-stats-go"
    "os"
)


func main() {
    emperor := penguin.NewClient()

    drops := []penguin.Drop{{
        DropType: penguin.NormalDrop,
        ItemID: "31001",
        Quantity: 2,
    },
    {
        DropType: penguin.NormalDrop,
        ItemID: "31002",
        Quantity: 1,
    }}

    ctx := context.Background()
    reportHash, err := emperor.ReportDrop(ctx, penguin.ServerUS, "main_01-01", drops, "myawesomeproject", "")
    if err != nil {
        fmt.Printf("err sending report to Penguin Stats: %v\n", err)
        os.Exit(1)
    }

    if err := emperor.RecallLastReport(ctx, reportHash, "myawesomeproject"); err != nil {
        fmt.Printf("err recalling last report: %v\n", err)
        os.Exit(1)
    }
}
```