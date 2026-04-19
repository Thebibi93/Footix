FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Image finale légère pour l'exécution
FROM alpine:latest

# Installe les certificats CA pour les requêtes HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copie le binaire compilé depuis l'étape précédente
COPY --from=builder /app/main .

# Expose le port 8080
EXPOSE 8080

# Commande de démarrage
CMD ["./main"]