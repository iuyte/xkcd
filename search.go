package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

var YoutubeKey = ""

const maxResults = 1

func YTSearch(query string) (string, error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: YoutubeKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		return "", errors.New("Error creating new YouTube client: " + err.Error())
	}

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(query).
		MaxResults(maxResults)
	response, err := call.Do()
	if err != nil {
		return "", errors.New("Error making search API call: " + err.Error())
	}

	return response.Items[0].Id.VideoId, nil
}

func DevKey() (token string) {
	token = os.Getenv("YOUTUBE_TOKEN")
	if (strings.Contains(token, "$") || token == "") && len(os.Args) > 1 {
		token = os.Args[1]
	} else {
		b, err := ioutil.ReadFile("/youtube.txt")
		if err != nil {
			fmt.Println(err)
		}
		token = strings.TrimSpace(strings.Trim(strings.Split(string(b), ";")[0], "\n"))
	}
	return
}
