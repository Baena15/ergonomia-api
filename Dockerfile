FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copiar todo el código
COPY . /app/

# Descargar dependencias y compilar
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copiar el binario
COPY --from=builder /app/api .

# Crear directorio web vacío si no existe
RUN mkdir -p /app/web

EXPOSE 8080
CMD ["./api"]
