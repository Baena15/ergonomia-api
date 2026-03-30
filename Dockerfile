FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instalar dependencias de build
RUN apk add --no-cache git

# Copiar dependencias primero (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api ./cmd/api

# Imagen final mínima
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copiar binario
COPY --from=builder /app/api .

# Copiar archivos estáticos
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./api"]
