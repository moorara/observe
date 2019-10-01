# log

This package provides **structured logging** for Go applications
(it is a wrapper for [go-kit logger](https://github.com/go-kit/kit/tree/master/log)).

Default output format is `log.JSON` and default log level is `log.InfoLevel`.

## Quick Start

You can use the **global/singelton** logger as follows:

```go
package main

import "github.com/moorara/observe/log"

func main() {
  log.SetOptions(log.Options{
    Name:        "service",
    Environment: "production",
    Region:      "us-east-1",
  })

  log.Infof("Hello, World!")
}
```

Output:

```json
{"caller":"main.go:12","environment":"production","level":"info","logger":"service","message":"Hello, World!","region":"us-east-1","timestamp":"2019-09-20T03:17:57.743345Z"}
```

Or you can create a new instance logger as follows:

```go
package main

import "github.com/moorara/observe/log"

func main() {
  logger := log.NewLogger(log.Options{
    Name:        "service",
    Environment: "production",
    Region:      "us-east-1",
    Level:       "debug",
    Format:      log.JSON,
  })

  logger = logger.With(
    "version", "0.1.0",
    "revision", "abcdef",
  )

  logger.DebugKV(
    "message", "Hello, World!",
    "requestId", "2222-bbbb",
  )
}
```

Output:

```json
{"caller":"main.go:19","environment":"production","level":"debug","logger":"service","message":"Hello, World!","region":"us-east-1","requestId":"2222-bbbb","revision":"abcdef","timestamp":"2019-09-20T03:25:50.124195Z","version":"0.1.0"}
```
