/**
 * Copyright (c) 2017 Ethan Wells
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
)

type ObjectResponse struct {
	Resp *http.Response
	Name string
}

type Streamer struct {
	Url       string
	GuildID   string
	ChannelID string
	S         *discordgo.Session
}

const (
	initZingMp3    string = "zing"
	initYoutube    string = "youtube"
	initSoundCloud string = "soundcloud"
	initFacebook   string = "facebook"
	initVimeo      string = "vimeo"
)

var (
	blocker = make(chan bool, 1)
	Stop    = make(map[string]bool)
	pause   = make(map[string]bool)
	exitQ   = false
	Streams = make(map[string]*Streamer)
)

func (s *Streamer) Stream() error {
	return Stream(s.Url, s.GuildID, s.ChannelID, s.S)
}

func UrlFromSearch(tag string) (string, error) {
	return YTSearch(tag)
}

func Stream(videoURL, guildID, channelID string, s *discordgo.Session) error {
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"

	videoInfo, err := ytdl.GetVideoInfo(videoURL)
	if err != nil {
		return err
	}

	formats := videoInfo.Formats.Best(ytdl.FormatAudioBitrateKey)
	if len(formats) < 1 {
		return errors.New("link error")
	}
	format := formats[0]
	downloadURL, err := videoInfo.GetDownloadURL(format)
	if err != nil {
		return err
	}

	encodingSession, err := dca.EncodeFile(downloadURL.String(), options)
	if err != nil {
		return err
	}

	defer func() {
		encodingSession.Cleanup()
		<-blocker
		Stop[guildID] = false
	}()
	blocker <- true
	if exitQ {
		return nil
	}

	var vc *discordgo.VoiceConnection
	defer func() {
		vc.Speaking(false)
		vc.Disconnect()
	}()
	vc, err = s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	vc.Speaking(true)
	done := make(chan error)
	session := dca.NewStream(encodingSession, vc, done)
	go func() {
		err = <-done
		Stop[guildID] = true
	}()
	for !Stop[guildID] {
		time.Sleep(250 * time.Millisecond)
		session.SetPaused(pause[guildID])
		vc.Speaking(pause[guildID])
	}
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
