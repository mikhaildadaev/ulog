---
outline: deep
---

# API / Запись в файл / Параметры

::: info **Информация**
На этой странице описаны все параметры конфигурации **SinkFile**: максимальный возраст файлов, количество резервных копий и размер файла перед ротацией. Каждый параметр показан с рабочим примером кода.
:::

## WithFileMaxAge
Устанавливает максимальное количество дней хранения старых лог-файлов. Файлы старше этого срока будут автоматически удалены при ротации.
```go
sinkFile, err := ulog.NewSinkFile("app.log",
    ulog.WithFileMaxAge(30),
)
if err != nil {
    fmt.Fprintf(ulog.DefaultWriterErr, "ulog: %v\n", err)
}
defer sinkFile.Close()
```

## WithFileMaxBackups
Устанавливает максимальное количество старых лог-файлов. При превышении этого лимита самые старые файлы удаляются при ротации.
```go
sinkFile, err := ulog.NewSinkFile("app.log",
    ulog.WithFileMaxBackups(10),
)
if err != nil {
    fmt.Fprintf(ulog.DefaultWriterErr, "ulog: %v\n", err)
}
defer sinkFile.Close()
```

## WithFileMaxSize
Устанавливает максимальный размер файла в мегабайтах перед ротацией. Когда текущий лог-файл превышает этот размер, он переименовывается, сжимается, и создаётся новый файл.
```go
sinkFile, err := ulog.NewSinkFile("app.log",
    ulog.WithFileMaxSize(100),
)
if err != nil {
    fmt.Fprintf(ulog.DefaultWriterErr, "ulog: %v\n", err)
}
defer sinkFile.Close()
```
