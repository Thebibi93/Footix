# Construction de l'application Go
FROM golang:1.25-alpine AS builder

# Définit le répertoire de travail dans le conteneur
WORKDIR /app

# Copie les fichiers de dépendances Go
COPY go.mod go.sum ./

# Télécharge les dépendances
RUN go mod download

# Copie le reste du code source
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