FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copiar archivos de módulos primero (cache)
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente explícitamente
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY migrations/ ./migrations/
COPY web/ ./web/

# Verificar estructura y compilar
RUN ls -la cmd/api/
RUN CGO_ENABLED=0 GOOS=linux go build -v -o api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/api .
RUN mkdir -p /app/web

EXPOSE 8080
CMD ["./api"]
