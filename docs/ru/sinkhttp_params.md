---
outline: deep
---

# API / Запись по сети / Параметры

::: info **Информация**
На этой странице описаны все параметры конфигурации `SinkHttp`: пакетная обработка, Circuit Breaker, дедупликация, повторные попытки, семплирование и другое. Каждый параметр показан со значением по умолчанию и кратким описанием.
:::

## WithHttpBatch
Пакетная отправка: до `size` сообщений или каждые `flushInterval`.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpBatch(100, 5*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpCircuitBreaker
Размыкание цепи после `maxFailures` ошибок, ожидание `timeout` перед восстановлением.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpCircuitBreaker(10, 10*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpDedupWindow
Игнорирование повторяющихся сообщений в течение `window`.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDedupWindow(5*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpDisabledBatch
Отключение пакетной отправки (отправлять немедленно).
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisabledBatch(),
)
defer sinkHttp.Close()
```

## WithHttpDisabledCircuit
Отключение Circuit Breaker.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisabledCircuit(),
)
defer sinkHttp.Close()
```

## WithHttpDisableKeepAlive
Отключение HTTP Keep-Alive соединений.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisableKeepAlive(),
)
defer sinkHttp.Close()
```

## WithHttpFilterData
Фильтрация по типу данных: `DataLog`, `DataMetric`, `DataTrace`.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpFilterData(ulog.DataLog),
)
defer sinkHttp.Close()
```

## WithHttpFilterLevel
Фильтрация по минимальному уровню: `LevelDebug`, `LevelError`, `LevelFatal`, `LevelInfo`, `LevelWarn`.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpFilterLevel(ulog.LevelError),
)
defer sinkHttp.Close()
```

## WithHttpFormatter
Пользовательская функция форматирования `func(attributes, fields) ([]byte, error)`.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
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
Добавление пользовательского HTTP-заголовка.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpHeader("Authorization", "Bearer token"),
    ulog.WithHttpHeader("X-Custom", "value"),
)
defer sinkHttp.Close()
```

## WithHttpMethod
HTTP-метод: `POST`, `PUT` и др.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpMethod("POST"),
)
defer sinkHttp.Close()
```

## WithHttpRetry
Повтор неудачных запросов до `maxRetries` раз с экспоненциальной задержкой `backoff`.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpRetry(3, time.Second),
)
defer sinkHttp.Close()
```

## WithHttpSampleRate
Семплирование 1 из `rate` сообщений для не-ошибочных уровней.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpSampleRate(100),
)
defer sinkHttp.Close()
```

## WithHttpSampleWindow
Сброс счётчика семплирования каждые `window`.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpSampleWindow(1*time.Minute),
)
defer sinkHttp.Close()
```

## WithHttpTimeout
Таймаут HTTP-клиента.
```go
import (
    "fmt"
    "github.com/mikhaildadaev/ulog"
)
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpTimeout(30*time.Second),
)
defer sinkHttp.Close()
```
