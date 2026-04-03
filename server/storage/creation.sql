DROP TABLE IF EXISTS Predictions CASCADE;
DROP TABLE IF EXISTS Matches CASCADE;
DROP TABLE IF EXISTS Teams CASCADE;
DROP TABLE IF EXISTS Leagues CASCADE;
DROP TABLE IF EXISTS UserPredictionHistory CASCADE;
DROP TABLE IF EXISTS UserScores CASCADE;
DROP TABLE IF EXISTS UserPredictions CASCADE;
DROP TABLE IF EXISTS Users CASCADE;
DROP TABLE IF EXISTS ChatMessages CASCADE;
DROP TABLE IF EXISTS ChatRoomCounters CASCADE;
DROP TABLE IF EXISTS ChatRooms CASCADE;
DROP TRIGGER IF EXISTS trg_init_chat_room_counter ON ChatRooms;
DROP FUNCTION IF EXISTS init_chat_room_counter();


CREATE TABLE IF NOT EXISTS Leagues (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(10) NOT NULL
);

CREATE TABLE IF NOT EXISTS Teams (
    id INT PRIMARY KEY, -- On utilise l'ID venant de l'API
    name VARCHAR(255) NOT NULL,
    short_name VARCHAR(50),
    crest_url TEXT
);

CREATE TABLE IF NOT EXISTS Matches (
    id SERIAL PRIMARY KEY,
    league_id INT NOT NULL,
    season INT NOT NULL,
    utc_date TIMESTAMP NOT NULL,
    home_team_id INT NOT NULL,
    away_team_id INT NOT NULL,
    home_score INT,
    away_score INT,
    status VARCHAR(20) NOT NULL,
    FOREIGN KEY (league_id) REFERENCES Leagues(id),
    FOREIGN KEY (home_team_id) REFERENCES Teams(id),
    FOREIGN KEY (away_team_id) REFERENCES Teams(id)
);

CREATE TABLE IF NOT EXISTS Predictions (
    match_id INT PRIMARY KEY,
    win_probability_home FLOAT NOT NULL,
    win_probability_away FLOAT NOT NULL,
    predicted_result VARCHAR(20) NOT NULL,
    FOREIGN KEY (match_id) REFERENCES Matches(id)
);

CREATE TABLE IF NOT EXISTS Users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS UserPredictions (
    user_id INT NOT NULL,
    match_id INT NOT NULL,
    predicted_result VARCHAR(20) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (match_id) REFERENCES Matches(id),
    PRIMARY KEY (user_id, match_id)
);

CREATE TABLE IF NOT EXISTS UserScores (
    user_id INT PRIMARY KEY,
    score INT NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES Users(id)
);

-- c'est bien que pour les match terminés, on puisse stocker le résultat réel du match dans la table UserPredictionHistory
-- et aussi pour un match donné, on peut trouver toutes les prédictions faites par les utilisateurs et les comparer au résultat réel du match
-- pour un user donné on lui trouve toutes les prédictions faites sur les match
CREATE TABLE IF NOT EXISTS UserPredictionHistory (
    user_id INT NOT NULL,
    match_id INT NOT NULL,
    predicted_result VARCHAR(20) NOT NULL,
    actual_result VARCHAR(20),
    prediction_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (match_id) REFERENCES Matches(id),
    PRIMARY KEY (user_id, match_id)
);



CREATE TABLE IF NOT EXISTS ChatRooms (
    id BIGSERIAL PRIMARY KEY,
    match_id INT NOT NULL,
    room_type VARCHAR(50) NOT NULL DEFAULT 'main',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_chatrooms_match
        FOREIGN KEY (match_id) REFERENCES Matches(id),

    CONSTRAINT uq_chatrooms_match_type
        UNIQUE (match_id, room_type)
);

-- Compteur de séquence par room.
-- Permet d'attribuer seq_in_room sans faire un MAX(...) à chaque insert.
CREATE TABLE IF NOT EXISTS ChatRoomCounters (
    chat_room_id BIGINT PRIMARY KEY,
    last_seq BIGINT NOT NULL DEFAULT 0,

    CONSTRAINT fk_chatroomcounters_room
        FOREIGN KEY (chat_room_id) REFERENCES ChatRooms(id) ON DELETE CASCADE
);

-- Messages stockés par room.
-- seq_in_room = ordre local dans le chat d'une room donnée.
CREATE TABLE IF NOT EXISTS ChatMessages (
    id BIGSERIAL PRIMARY KEY,
    chat_room_id BIGINT NOT NULL,
    seq_in_room BIGINT NOT NULL,
    user_id INT NOT NULL,
    message VARCHAR(1024) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_chatmessages_room
        FOREIGN KEY (chat_room_id) REFERENCES ChatRooms(id) ON DELETE CASCADE,

    CONSTRAINT fk_chatmessages_user
        FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE,

    CONSTRAINT uq_chatmessages_room_seq
        UNIQUE (chat_room_id, seq_in_room)
);

-- Index principal pour le rattrapage incrémental :
-- "donne-moi les messages de cette room après telle séquence"
CREATE INDEX IF NOT EXISTS idx_chatmessages_room_seq
ON ChatMessages (chat_room_id, seq_in_room);

-- Index utile pour retrouver rapidement la room principale d'un match
CREATE INDEX IF NOT EXISTS idx_chatrooms_match
ON ChatRooms (match_id);

-- ============================================================
-- ========== AUTO-INIT DU COMPTEUR LORS D'UNE ROOM ===========
-- ============================================================

CREATE OR REPLACE FUNCTION init_chat_room_counter()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO ChatRoomCounters(chat_room_id, last_seq)
    VALUES (NEW.id, 0)
    ON CONFLICT (chat_room_id) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_init_chat_room_counter
AFTER INSERT ON ChatRooms
FOR EACH ROW
EXECUTE FUNCTION init_chat_room_counter();


-- ============================================================
-- ===================== EXEMPLES D'USAGE =====================
-- ============================================================

-- 1) Obtenir la room principale d'un match
-- Usage :
--   SELECT id FROM ChatRooms WHERE match_id = 123 AND room_type = 'main';

-- 2) Ajouter un message de manière atomique
--    (on incrémente d'abord la séquence de la room, puis on insère)
--
-- Exemple paramétré :
--   $1 = chat_room_id
--   $2 = user_id
--   $3 = message
--
-- WITH next_seq AS (
--     UPDATE ChatRoomCounters
--     SET last_seq = last_seq + 1
--     WHERE chat_room_id = $1
--     RETURNING last_seq
-- )
-- INSERT INTO ChatMessages (chat_room_id, seq_in_room, user_id, message)
-- SELECT $1, next_seq.last_seq, $2, $3
-- FROM next_seq
-- RETURNING id, chat_room_id, seq_in_room, user_id, message, created_at;

-- 3) Charger uniquement les nouveaux messages d'une room
-- Usage :
--   roomId = 42
--   afterSeq = 99
--
-- SELECT id, chat_room_id, seq_in_room, user_id, message, created_at
-- FROM ChatMessages
-- WHERE chat_room_id = $1
--   AND seq_in_room > $2
-- ORDER BY seq_in_room ASC
-- LIMIT 100;

-- 4) Charger les 50 derniers messages d'une room lors de l'arrivée sur le chat
--
-- SELECT id, chat_room_id, seq_in_room, user_id, message, created_at
-- FROM ChatMessages
-- WHERE chat_room_id = $1
-- ORDER BY seq_in_room DESC
-- LIMIT 50;