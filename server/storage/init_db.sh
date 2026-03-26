#!/bin/bash

# 1. Récupère le chemin absolu du dossier où se trouve ce script
# Peu importe où le script est lancé, SCRIPT_DIR sera toujours .../Footix/server/storage car
#${BASH_SOURCE[0]} : C'est une variable interne à Bash qui contient le chemin du script en cours d'exécution.
#dirname : extrait la partie "dossier" d'un chemin. Si le chemin est
# /Footix/server/storage/init_db.sh, dirname renvoie /Footix/server/storage.
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

# 2. Déduit les chemins relatifs à partir de SCRIPT_DIR
# On remonte d'un cran (..) pour sortir de storage/ et aller dans resources/
PROP_FILE="$SCRIPT_DIR/../resources/properties.txt"

# Le fichier SQL est dans le même dossier que le script
SQL_FILE="$SCRIPT_DIR/creation.sql"

# 3. Optionnel : Trouver la racine du projet (Footix/)
PROJECT_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)

echo "--- Diagnostic des chemins ---"
echo "Script situé dans : $SCRIPT_DIR"
echo "Racine du projet  : $PROJECT_ROOT"
echo "Fichier config    : $PROP_FILE"
echo "------------------------------"

# Vérification de l'existence du fichier de propriétés
if [ ! -f "$PROP_FILE" ]; then
    echo "Erreur : Le fichier de configuration $PROP_FILE est introuvable."
    exit 1
fi

# Fonction pour extraire une valeur du fichier properties.txt
#^ : Signifie "commence par   arg=   " avec arg = db_user, db_password, etc. envoyé à la fonction get_prop
# puis |  : Envoie le résultat de grep à la commande suivante.
# -d'=' : Définit le "délimiteur" comme étant le signe égal.
# -f2 : Demande de garder le deuxième champ (ce qui est à droite du égal). (la valeur de la propriété)
get_prop() {
    grep "^$1=" "$PROP_FILE" | cut -d'=' -f2
}

# Extraction des variables
DB_NAME=$(get_prop "db_name")
DB_USER=$(get_prop "db_user")
DB_PASS=$(get_prop "db_password")
DB_HOST=$(get_prop "db_host")

# Vérification si les variables essentielles sont remplies
if [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASS" ]; then
    echo "Erreur : Certaines propriétés sont manquantes dans $PROP_FILE"
    exit 1
fi

# Export du mot de passe pour psql (évite la saisie manuelle)
export PGPASSWORD=$DB_PASS

echo "--- Initialisation de la base de données ---"
echo "Base : $DB_NAME"
echo "Utilisateur : $DB_USER"
echo "Hôte : $DB_HOST"

# Exécution du script SQL
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f "$SQL_FILE"

# Vérification du résultat
if [ $? -eq 0 ]; then
    echo "--------------------------------------------"
    echo "Succès : Les tables ont été créées."
else
    echo "--------------------------------------------"
    echo "Erreur lors de l'exécution du script SQL."
fi

# Nettoyage
unset PGPASSWORD