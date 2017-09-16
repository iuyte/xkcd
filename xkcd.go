package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type XKCD struct {
	Alt        string `json:"alt"`
	Day        string `json:"day"`
	Img        string `json:"img"`
	Link       string `json:"link"`
	Month      string `json:"month"`
	News       string `json:"news"`
	Num        int    `json:"num"`
	SafeTitle  string `json:"safe_title"`
	Title      string `json:"title"`
	Transcript string `json:"transcript"`
	Year       string `json:"year"`
}

func GetXkcd(num string) (xkcd XKCD, e error) {
	var (
		r    *http.Response
		data []byte
	)

	r, e = http.Get("https://xkcd.com/" + num + "/info.0.json")
	if e != nil {
		return xkcd, e
	}

	defer r.Body.Close()
	data, e = ioutil.ReadAll(r.Body)
	if e != nil {
		return xkcd, e
	}

	e = json.Unmarshal(data, &xkcd)
	if e != nil {
		return xkcd, e
	}

	return xkcd, nil
}

const prefix string = ";"

var (
	token string
)

func main() {
	token = Token()
	if token == "" {
		fmt.Println("No token provided. Please place a token followed by a ';' at ../token.txt")
		return
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	fmt.Println("The bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
	os.Exit(0)
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == "315552571823489024" {
			_, _ = s.ChannelMessageSend(channel.ID, "Connected.")
			return
		}
	}
}

func Token() string {
	b, err := ioutil.ReadFile("./token.txt")
	if err != nil {
		fmt.Println(err)
	}
	c := string(b)
	return strings.Split(c, ";")[0]
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, "Always watching...")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Message.Content) < 1 {
		return
	}
	if m.Message.ContentWithMentionsReplaced()[:len(prefix)] != prefix {
		return
	}

	c := strings.Split(strings.TrimSpace(strings.TrimPrefix(m.Message.Content, prefix)), " ")

	if c[0] == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong!")
		return
	}

	if c[0] == "xkcd" {
		if len(c) < 2 {
			c = append(c, "")
		}

		xkcd, err := GetXkcd(c[1])
		if err != nil {
			fmt.Println(err)
			return
		}

		e := &discordgo.MessageEmbed{
			Title: "xkcd #" + strconv.Itoa(xkcd.Num) + ": " + xkcd.Title,
			URL:   "https://xkcd.com/" + strconv.Itoa(xkcd.Num),
			Color: 7506394,
			Type:  "rich",
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
			Fields: {&discordgo.MessageEmbedField{
				Value: xkcd.Alt,
			}},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		_, err = s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
		} else {
			return
		}

		s.ChannelMessageSend(m.ChannelID, xkcd.Img)
	}
}
