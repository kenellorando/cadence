SHOW databases;
CREATE DATABASE cadence;
USE cadence;

CREATE TABLE music (
	song_id INTEGER NOT NULL AUTO_INCREMENT,
    song_title VARCHAR(130) NOT NULL,
    album_title VARCHAR(130),
    artist_name VARCHAR(80), 
    song_length FLOAT NOT NULL,
    song_path VARCHAR(250) UNIQUE NOT NULL,
PRIMARY KEY (song_id, song_path)
);

/*
EXPLAIN music;
SELECT * FROM music;
*/

/*
CREATE USER 'populator'@'localhost' IDENTIFIED BY 'populator1';
GRANT ALL ON cadence.* TO 'populator'@'localhost';
FLUSH PRIVILEGES;
*/

/*
DROP USER 'populator'@'localhost';
DROP TABLE music;
*/