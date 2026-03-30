# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instalar git (necesario para go mod)
RUN apk add --no-cache git

# Copiar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api ./cmd/api

# Final stage - imagen mínima
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copiar binario
COPY --from=builder /app/api .

# Copiar archivos estáticos
COPY --from=builder /app/web ./web

# Puerto expuesto
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Comando de inicio
CMD ["./api"]
