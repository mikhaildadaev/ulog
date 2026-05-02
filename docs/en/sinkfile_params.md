---
outline: deep
---

# API / SinkFile / Params

::: info **Info**
This page covers all configuration options for **SinkFile**: maximum file age, backup count, and file size before rotation. Each option is shown with a working code example and expected behavior.
:::

## WithFileMaxAge
Sets the maximum number of days to keep old log files. Files older than this will be automatically deleted during rotation.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkFile, err := ulog.NewSinkFile("app.log",
    ulog.WithFileMaxAge(30),
)
if err != nil {
    fmt.Fprintf(ulog.DefaultWriterErr, "ulog: %v\n", err)
}
defer sinkFile.Close()
```

## WithFileMaxBackups
Sets the maximum number of old log files to keep. When this limit is exceeded, the oldest backup files are deleted during rotation.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkFile, err := ulog.NewSinkFile("app.log",
    ulog.WithFileMaxBackups(10),
)
if err != nil {
    fmt.Fprintf(ulog.DefaultWriterErr, "ulog: %v\n", err)
}
defer sinkFile.Close()
```

## WithFileMaxSize
Sets the maximum file size in megabytes before rotation is triggered. When the current log file exceeds this size, it is renamed, compressed, and a new file is started.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkFile, err := ulog.NewSinkFile("app.log",
    ulog.WithFileMaxSize(100),
)
if err != nil {
    fmt.Fprintf(ulog.DefaultWriterErr, "ulog: %v\n", err)
}
defer sinkFile.Close()
```