package storage

import (
	"database/sql"
	"fmt"
	"os"
	"bufio"
	"strings"
	_ "github.com/lib/pq" // L'underscore est crucial : on importe le driver pour ses effets de bord (init)
)
/*
	On encapsule la base de données dans une structure pour pouvoir lui attacher des méthodes.
*/

// Config stocke nos paramètres du fichier properties.txt
type Config struct {
	DBName     string
	DBUser     string
	DBPassword string
	DBHost     string
	APIToken   string // token de l'API football-data.org
}

// InitDB charge la config et ouvre la connexion
func InitDB() (*sql.DB, error) {
	config, err := LoadProperties("resources/properties.txt")
	if err != nil {
		return nil, err
	}

	// Création de la chaîne de connexion (DSN) <=> data source name
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName)

	// sql.Open ne crée pas de connexion immédiate, il vérifie juste les arguments
	// DOC : func sql.Open(driverName string, dataSourceName string) (*sql.DB, error)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// db.Ping() est INDISPENSABLE pour vérifier que la connexion est réellement établie
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// On parse les porprietées externalisées depuis le fichier resources/properties.txt
func LoadProperties(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		switch key {
			case "db_name": config.DBName = value
			case "db_user": config.DBUser = value
			case "db_password": config.DBPassword = value
			case "db_host": config.DBHost = value
			case "token": config.APIToken = value 
		}
	}
	return config, scanner.Err()
}