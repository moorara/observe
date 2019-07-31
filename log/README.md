# log

This package provides **structured logging** for Go applications
(it is a wrapper for [go-kit/kit/log](https://github.com/go-kit/kit/tree/master/log)).

Default output format is `log.JSON` and default log level is `log.InfoLevel`.

## Quick Start

You can use the **global/singelton** logger as follows:

```go
package main

import (
  "errors"

  "github.com/moorara/observe/log"
)

func main() {
  log.SetOptions(log.Options{
    Environment: "staging",
    Region:      "us-east-1",
  })

  log.Error(
    "message", "Hello, World!",
    "error", errors.New("too late!"),
  )
}
```

Output:

```json
{"caller":"main.go:15","environment":"staging","error":"too late!","level":"error","message":"Hello, World!","region":"us-east-1","timestamp":"2019-07-31T05:11:42.19237Z"}
```

Or you can create a new instance logger as follows:

```go
package main

import "github.com/moorara/observe/log"

func main() {
  logger := log.NewLogger(log.Options{
    Format:      log.JSON,
    Level:       "debug",
    Name:        "hello-world",
    Environment: "production",
    Region:      "us-east-1",
    Component:   "main-function",
  })

  logger.Debug(
    "message", "Hello, World!",
    "context", map[string]interface{}{
      "retries": 4,
    },
  )
}
```

Output:

```json
{"caller":"main.go:15","component":"main-function","context":{"retries":4},"environment":"production","level":"debug","logger":"hello-world","message":"Hello, World!","region":"us-east-1","timestamp":"2019-07-31T05:12:38.585191Z"}
```
