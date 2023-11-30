CREATE TABLE abbreviations (
    id SERIAL PRIMARY KEY,
    abbreviation TEXT NOT NULL,
    meaning TEXT NOT NULL,
    author TEXT
);