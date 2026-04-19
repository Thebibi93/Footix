package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // L'underscore est crucial : on importe le driver pour ses effets de bord (init)
)

/*
	On encapsule la base de données dans une structure pour pouvoir lui attacher des méthodes.
*/

// Config stocke nos paramètres du fichier env
type Config struct {
	DBName     string
	DBUser     string
	DBPassword string
	DBHost     string
	APIToken   string // token de l'API football-data.org
}

func LoadConfig() (Config, error) {
	cfg := Config{
		DBName:     getEnv("DB_NAME", ""),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "localhost"),
		APIToken:   getEnv("API_TOKEN", ""),
	}

	if cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.APIToken == "" {
		return cfg, fmt.Errorf("configuration incomplète : DB_NAME, DB_USER, DB_PASSWORD et API_TOKEN sont requis")
	}

	return cfg, nil
}

// getEnv récupère une variable d'environnement avec une valeur par défaut
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitDB charge la config et ouvre la connexion
func InitDB() (*sql.DB, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Création de la chaîne de connexion (DSN) <=> data source name
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=require",
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
