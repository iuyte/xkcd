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

var YoutubeKey = "https://developers.google.com"

const maxResults = 50

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

	// Group video, channel, and playlist results in separate lists.
	videos := make(map[string]string)

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = item.Snippet.Title
		default:
			continue
		}
	}

	for _, item := range response.Items {
		if strings.Contains(item.Id.Kind, "video") {
			return item.Id.VideoId, nil
		}
	}

	return "", errors.New("Video not found")
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
