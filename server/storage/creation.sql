DROP TABLE IF EXISTS Predictions;
DROP TABLE IF EXISTS Matches;
DROP TABLE IF EXISTS Teams;
DROP TABLE IF EXISTS Leagues;


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