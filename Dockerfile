# Сборка потокового парсера GAR XML
FROM golang:1.21 AS builder
WORKDIR /app

# Сначала загрузите зависимости для лучшего кэширования слоёв
COPY go.mod ./
RUN go mod download

# Копирование остального исходного кода
COPY . .

# Сборка статически связанного бинарного файла для небольшого образа выполнения
# Компилируйте только точку входа CLI, чтобы избежать ошибок "несколько пакетов"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/fias-parser ./cmd/fias_parser

FROM debian:bookworm-slim
WORKDIR /
COPY --from=builder /bin/fias-parser /usr/local/bin/
COPY gar_schemas /gar_schemas

ENTRYPOINT ["/usr/local/bin/fias-parser"]
CMD ["--help"]
