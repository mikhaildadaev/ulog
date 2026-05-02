---
outline: deep
---

# API / 文件接收器 / 参数

::: info **关于**
本页涵盖了 **SinkFile** 的所有配置选项：最大文件保留天数、备份文件数量和文件大小限制。每个选项都附有可运行的代码示例。
:::

## WithFileMaxAge
设置旧日志文件的最大保留天数。超过此期限的文件将在轮换时自动删除。
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
设置旧日志文件的最大保留数量。超过此限制时，最旧的备份文件将在轮换时删除。
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
设置触发轮换的最大文件大小（以兆字节为单位）。当前日志文件超过此大小时，将被重命名、压缩，并创建新文件。
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
