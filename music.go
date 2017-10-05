package main

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dwarvesf/glod"
	"github.com/dwarvesf/glod/chiasenhac"
	"github.com/dwarvesf/glod/facebook"
	"github.com/dwarvesf/glod/soundcloud"
	"github.com/dwarvesf/glod/vimeo"
	"github.com/dwarvesf/glod/youtube"
	"github.com/dwarvesf/glod/zing"
	"github.com/jonas747/dca"
)

const (
	initZingMp3    string = "zing"
	initYoutube    string = "youtube"
	initSoundCloud string = "soundcloud"
	initChiaSeNhac string = "chiasenhac"
	initFacebook   string = "facebook"
	initVimeo      string = "vimeo"
)

type ObjectResponse struct {
	Resp *http.Response
	Name string
}

func Stream(link string, s *discordgo.Session, guildID, channelID string) error {
	var ggl glod.Source
	if ggl = func() glod.Source {
		switch {
		case strings.Contains(link, initZingMp3):
			return &zing.Zing{}
		case strings.Contains(link, initYoutube):
			return &youtube.Youtube{}
		case strings.Contains(link, initSoundCloud):
			return &soundcloud.SoundCloud{}
		case strings.Contains(link, initChiaSeNhac):
			return &chiasenhac.ChiaSeNhac{}
		case strings.Contains(link, initFacebook):
			return &facebook.Facebook{}
		case strings.Contains(link, initVimeo):
			return &vimeo.Vimeo{}
		}
		return nil
	}(); ggl == nil {
		return errors.New("Cannot read source link")
	}

	var objs []ObjectResponse
	listResponse, err := ggl.GetDirectLink(link)
	if err != nil {
		return err
	}
	for _, r := range listResponse {
		temp := r.StreamURL
		if strings.Contains(link, initYoutube) || strings.Contains(link, initZingMp3) || strings.Contains(link, initVimeo) {
			splitUrl := strings.Split(temp, "~")
			temp = splitUrl[0]
		}

		resp, err := http.Get(temp)
		if err != nil {
			return errors.New("failed to get response from  stream")
		}

		fullName := fmt.Sprintf("%s%s", r.Title, ".mp3")

		fullName = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, fullName)

		objs = append(objs, ObjectResponse{
			resp,
			fullName,
		})
	}

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	time.Sleep(250 * time.Millisecond)
	vc.Speaking(true)

	for _, o := range objs {
		defer o.Resp.Body.Close()

		rd := o.Resp.Body
		decoder := dca.NewDecoder(rd)

		for {
			frame, err := decoder.OpusFrame()
			if err != nil {
				if err != io.EOF {
					return errors.New("connection bork 2")
				}

				break
			}

			select {
			case vc.OpusSend <- frame:
			case <-time.After(time.Second):
				return errors.New("connection bork 2")
			}
		}
	}
	vc.Speaking(false)
	time.Sleep(250 * time.Millisecond)
	vc.Disconnect()
	return nil
}
