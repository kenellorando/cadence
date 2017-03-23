/* List of example commands for a music database
supports song titles, artist names, and file paths
Tested for LASTORDER's database*/
    
/* DB CREATE */
SHOW databases;
CREATE DATABASE music;
USE music;

/* DROPS */
DROP TABLE songs;

/* TABLE CREATION */
CREATE TABLE songs (
    song_id INTEGER NOT NULL AUTO_INCREMENT,
    song_title VARCHAR(100) NOT NULL,
    artist_name VARCHAR(50) DEFAULT NULL,
    song_path VARCHAR(100) NOT NULL,
    PRIMARY KEY (song_id)
);

/* DESCRIBE TABLES */
EXPLAIN songs;

/* INSERTS, note the double backslash*/
INSERT INTO songs (song_title, artist_name, song_path)
VALUES ("database", "MAN WITH A MISSION", "C:\\Users\\kenel\\Music\\MAN WITH A MISSION\\01 - database feat.TAKUMA (10-FEET).mp3");

INSERT INTO songs (song_title, artist_name, song_path)
VALUES ("Hello,world!", "BUMP OF CHICKEN", "C:\\Users\\kenel\\Music\\BUMP OF CHICKEN\\01. Hello,world!.mp3");


/* SELECTS */
SELECT 
    *
FROM
    songs;
    
SELECT 
    *
FROM
    songs
WHERE
    artist_name = 'BUMP OF CHICKEN';
    
SELECT 
    *
FROM
    songs
WHERE
    song_title = 'database';