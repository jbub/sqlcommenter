# sqlcommenter
[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/jbub/sqlcommenter)
[![Build Status](https://cloud.drone.io/api/badges/jbub/sqlcommenter/status.svg)](https://cloud.drone.io/jbub/sqlcommenter)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbub/sqlcommenter)](https://goreportcard.com/report/github.com/jbub/sqlcommenter)

Go implementation of https://google.github.io/sqlcommenter/.

## Usage with pgx stdlib driver

```go
package main

import (
    "context"
    "database/sql"

    "github.com/jackc/pgx/v4/stdlib"
    "github.com/jbub/sqlcommenter"
)

type contextKey int

const contextKeyUserID contextKey = 0

func withUserID(ctx context.Context, key string) context.Context {
    return context.WithValue(ctx, contextKeyUserID, key)
}

func userIDFromContext(ctx context.Context) string {
    return ctx.Value(contextKeyUserID).(string)
}

func main() {
    pgxDrv := stdlib.GetDefaultDriver()
    drv := sqlcommenter.WrapDriver(pgxDrv,
        sqlcommenter.WithAttrPairs("application", "hello-app"),
        sqlcommenter.WithAttrFunc(func(ctx context.Context) sqlcommenter.Attrs {
            return sqlcommenter.AttrPairs("user-id", userIDFromContext(ctx))
        }),
    )

    sql.Register("pgx-sqlcommenter", drv)

    db, err := sql.Open("pgx-sqlcommenter", "postgres://user@host:5432/db")
    if err != nil {
        // handle error
    }
    defer db.Close()
    
    ctx := context.Background()

    rows, err := db.QueryContext(withUserID(ctx, "22"), "SELECT 1")
    if err != nil {
        // handle error
    }
    defer rows.Close()
    
    // will produce the following query: SELECT 1 /* application='hello-app',user-id='22' */
}
```