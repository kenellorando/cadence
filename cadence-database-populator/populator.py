from tinytag import TinyTag
import os
import mysql.connector
from mysql.connector import MySQLConnection, Error

# Path of directory holding music
path = "C:/Users/kenel/Documents/Code/cadence-database-populator/test-music"

# Database configuration stored here
config = {
    'user': 'populator',
    'password': 'populator1',
    'host': '127.0.0.1',
    'database': 'cadence',
}


# Connection grabs configuration above
connection = mysql.connector.connect(**config)
cursor = connection.cursor()

# Loop through each file in given path
print("BEGINNING READ OF CONTENTS IN %s" % path)
for filename in os.listdir(path):
    print("")
    if filename.endswith(".mp3") or filename.endswith(".m4u") or filename.endswith(".flac"):
        # Set song path String and print it
        song_path = os.path.join(path, filename)
        tag = TinyTag.get(song_path)
        print("SONG PATH: " + song_path)

        # Start setting metadata here
        # Song title
        song_title = tag.title
        print("SONG TITLE: %s" % song_title)
        # Album title
        album_title = tag.album
        print("ALBUM TITLE: %s" % album_title)
        # Artist name
        artist_name = tag.artist
        print("ARTIST NAME: %s" % artist_name)
        # Song length in seconds
        song_length = tag.duration
        print("SONG LENGTH: %s" % song_length)

        # Insertion statement
        insert_statement = (
            "INSERT INTO cadence.music (song_title, album_title, artist_name, song_length, song_path) VALUES (%s, %s, %s, %s, %s)")
        # Values to be inserted
        song_values = (song_title, album_title, artist_name, song_length, song_path)



        # Try each insert
        try:
            cursor.execute(insert_statement, song_values)
            connection.commit()
            print("Entry inserted!")
        except Error as error:
            print("Did not insert! %s" % error)

        continue
    else:
        continue

# Closeouts
cursor.close()
connection.close()
