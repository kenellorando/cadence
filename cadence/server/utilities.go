// utilities.go
// Utility functions used by handlers

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/dhowden/tag"
	"github.com/kenellorando/clog"
)

func startsWith(str string, prefix string) bool {
	return len(str) >= len(prefix) && strings.EqualFold(str[:len(prefix)], prefix)
}

func endsWith(str string, suffix string) bool {
	return len(str) >= len(suffix) && strings.EqualFold(str[len(str)-len(suffix):], suffix)
}

func tokenCheck(token string) bool {
	clog.Info("tokenCheck", fmt.Sprintf("Checking token %s...", token))

	if len(token) != 26 {
		clog.Debug("tokenCheck", fmt.Sprintf("Token %s does not satisfy length requirements.", token))
		return false
	}

	// Check the whitelist. If this fails, the whitelist is not configured. No panic is thrown, but the bypass is denied.
	b, err := ioutil.ReadFile(c.WhitelistPath)
	if err != nil {
		return false
	}
	s := string(b)

	if strings.Contains(s, token) {
		clog.Info("tokenCheck", fmt.Sprintf("Token %s is valid.", token))
		return true
	}
	clog.Info("tokenCheck", fmt.Sprintf("Token %s is invalid.", token))
	return false
}

// Queries the metadata DB for the path of a given song, reads the file data for artwork, then returns artwork directly.
func getAlbumArt(currentArtist string, currentTitle string) []byte {
	log.Printf("%s %s", currentArtist, currentTitle)

	selectStatement := fmt.Sprintf("SELECT path FROM %s WHERE artist=\"%v\" AND title=\"%v\";", c.MetadataTable, currentArtist, currentTitle)
	rows, err := database.Query(selectStatement)
	if err != nil {
		clog.Error("getAlbumArt", "Could not query the DB for a path.", err)
		return nil
	}
	if err != nil {
		clog.Error("getAlbumArt", "Could not query the DB for a path.", err)
		return nil
	}

	var pic []byte

	for rows.Next() {
		var path string
		err := rows.Scan(&path)
		if err != nil {
			clog.Debug("getAlbumArt", "ae")
			return nil
		}
		// Open a file for reading
		file, err := os.Open(path)
		if err != nil {
			clog.Error("getAlbumArt", "Could not open music file for album art.", err)
			return nil
		}
		// Read metadata from the file
		tags, err := tag.ReadFrom(file)
		if err != nil {
			clog.Error("getAlbumArt", "Could not read tags from file for album art.", err)
			return nil
		}
		pic = tags.Picture().Data
	}
	return pic
}
