---
outline: deep
---

# API / HTTP 接收器 / 参数

::: info **关于**
本页涵盖了 `SinkHttp` 的所有配置选项：批处理、断路器、去重、重试、采样等。每个选项都附有其默认值和简要描述。
:::

## WithHttpBatch
批量发送消息：最多 `size` 条消息或每 `flushInterval` 发送一次。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpBatch(100, 5*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpCircuitBreaker
在 `maxFailures` 次错误后断开电路，等待 `timeout` 后恢复。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpCircuitBreaker(10, 10*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpDedupWindow
在 `window` 时间内忽略重复消息。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDedupWindow(5*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpDisabledBatch
禁用消息批处理（立即发送）。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisabledBatch(),
)
defer sinkHttp.Close()
```

## WithHttpDisabledCircuit
禁用断路器。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisabledCircuit(),
)
defer sinkHttp.Close()
```

## WithHttpDisableKeepAlive
禁用 HTTP Keep-Alive 连接。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisableKeepAlive(),
)
defer sinkHttp.Close()
```

## WithHttpFilterData
按数据类型过滤：`DataLog`、`DataMetric`、`DataTrace`。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpFilterData(ulog.DataLog),
)
defer sinkHttp.Close()
```

## WithHttpFilterLevel
按最低级别过滤：`LevelDebug`、`LevelError`、`LevelFatal`、`LevelInfo`、`LevelWarn`。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpFilterLevel(ulog.LevelError),
)
defer sinkHttp.Close()
```

## WithHttpFormatter
自定义格式化函数 `func(attributes, fields) ([]byte, error)`。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
        // Custom formatting logic
        return json.Marshal(fields)
    }),
)
defer sinkHttp.Close()
```

## WithHttpHeader
添加自定义 HTTP 头。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpHeader("Authorization", "Bearer token"),
    ulog.WithHttpHeader("X-Custom", "value"),
)
defer sinkHttp.Close()
```

## WithHttpMethod
HTTP 方法：`POST`、`PUT` 等。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpMethod("POST"),
)
defer sinkHttp.Close()
```

## WithHttpRetry
重试失败的请求，最多 `maxRetries` 次，采用指数退避 `backoff`。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpRetry(3, time.Second),
)
defer sinkHttp.Close()
```

## WithHttpSampleRate
对非错误级别的消息进行采样，采样率为 1/`rate`。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpSampleRate(100),
)
defer sinkHttp.Close()
```

## WithHttpSampleWindow
每 `window` 重置采样计数器。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpSampleWindow(1*time.Minute),
)
defer sinkHttp.Close()
```

## WithHttpTimeout
HTTP 客户端超时。
```go
import (
    "fmt"
    "github.com/mikhaildadaev/uuid"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpTimeout(30*time.Second),
)
defer sinkHttp.Close()
```
