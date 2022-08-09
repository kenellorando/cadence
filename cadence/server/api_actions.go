// api_actions.go
// Repeatable actions used by API functions. Mostly database and audio client functions.
package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/kenellorando/clog"
)

// Song file metadata object
type SongData struct {
	ID     int
	Artist string
	Title  string
	Album  string
	Genre  string
	Year   int
}

// Takes a search string, returns a slice of SongData of songs by relevance.
func dbQuery(query string) (queryResults []SongData, err error) {
	clog.Info("dbQuery", fmt.Sprintf("Searching database for query: '%v'", query))

	selectWhereStatement := fmt.Sprintf("SELECT \"rowid\", \"artist\", \"title\",\"album\", \"genre\", \"year\" FROM %s ", c.MetadataTable) + "WHERE artist LIKE $1 OR title LIKE $2 ORDER BY rank"
	rows, err := db.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%")
	if err != nil {
		clog.Error("Search", "Database search failed.", err)
		return nil, err
	}

	for rows.Next() {
		song := new(SongData)
		err = rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
		if err != nil {
			clog.Error("Search", "Data scan failed.", err)
			return
		}
		queryResults = append(queryResults, SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year})
	}

	return queryResults, nil
}

// Takes a song ID integer, returns a string absolute path of the song.
// Separated in order to keep filesystem paths internal.
func dbQueryPath(id int) (path string, err error) {
	clog.Info("dbQuery", fmt.Sprintf("Searching database for the path of song: '%v'", id))

	selectWhereStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE rowid=%v", c.MetadataTable, id)

	rows, err := db.Query(selectWhereStatement)
	if err != nil {
		clog.Error("Search", "Database search failed.", err)
		return "", err
	}
	for rows.Next() {
		err = rows.Scan(&path)
		if err != nil {
			clog.Error("Search", "Data scan failed.", err)
			return "", err
		}
	}

	return path, nil
}

// Takes an absolute song path, submits the path to be queued in Liquidsoap.
// Returns the response message from Liquidsoap.
func pushRequest(path string) (message string, err error) {
	// Telnet to liquidsoap
	clog.Debug("Request", "Connecting to liquidsoap service...")
	conn, err := net.Dial("tcp", c.SourceAddress+c.SourcePort)
	defer conn.Close()
	if err != nil {
		clog.Error("Request", "Failed to connect to audio source server.", err)
		return "", err
	}

	// Push song request to source service
	fmt.Fprintf(conn, "request.push "+path+"\n")
	// Listen for response
	message, _ = bufio.NewReader(conn).ReadString('\n')
	clog.Info("Request", fmt.Sprintf("Message from audio source server: %s", message))
	// Goodbye
	fmt.Fprintf(conn, "quit"+"\n")

	return message, nil
}
