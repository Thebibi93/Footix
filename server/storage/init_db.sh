#!/bin/bash

#!/bin/bash

# Charge automatiquement le .env s'il existe à la racine du projet
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
ENV_FILE="$SCRIPT_DIR/../../.env"
if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
    echo "Fichier .env chargé depuis $ENV_FILE"
fi

# Ensuite la vérification des variables
if [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ]; then
    echo "Erreur : Les variables DB_NAME, DB_USER et DB_PASSWORD doivent être définies."
    exit 1
fi

DB_HOST=${DB_HOST:-localhost}

# Chemin absolu vers le script SQL (toujours dans le même dossier que ce script)
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
SQL_FILE="$SCRIPT_DIR/creation.sql"

if [ ! -f "$SQL_FILE" ]; then
    echo "Fichier SQL introuvable : $SQL_FILE"
    exit 1
fi

export PGPASSWORD="$DB_PASSWORD"

echo "Initialisation de la base '$DB_NAME' sur $DB_HOST..."
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f "$SQL_FILE"

if [ $? -eq 0 ]; then
    echo "Tables créées avec succès."
else
    echo "Erreur lors de l'exécution du script SQL."
fi

unset PGPASSWORD