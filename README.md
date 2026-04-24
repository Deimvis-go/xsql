# sql

Go library with SQL related solutions.

## Packages

* [`sqlfw`](./sqlfw) — struct-based query manager. Declare queries as struct fields, load them from an `embed.FS`, look up metadata by field pointer.

## Quick Start

```golang
package quick_start

import (
    "embed"

    "github.com/Deimvis-go/sql/sqlfw"
)

//go:embed queries
var queries embed.FS

type Queries struct {
    SelectUser string `sql:"path=select_user.sql;name=select_user"`
    InsertUser string `sql:"path=insert_user.sql;name=insert_user"`
}

func main() {
    qm := sqlfw.NewStructQueryManager[Queries]()
    _, _ = qm.ReadFromFS(queries, "queries")
    _ = qm.Queries().SelectUser
}
```
