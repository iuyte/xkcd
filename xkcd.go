package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/http"
)

const prefix string = ">>"

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
		w io.ReadWriter
		b []byte = make([]byte, 0)
	)
	_, e = http.Get(w, strings.Join([]string{"https://xkcd.com/", num, "info.0.json"}, ""))
	if e != nil {
		return xkcd, e
	}
	_, e = w.Read(b)
	if e != nil {
		return xkcd, e
	}
	json.Unmarshal(b, xkcd)
	return xkcd, e
}

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
	b, err := ioutil.ReadFile("/token.txt")
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
	if m.Message.ContentWithMentionsReplaced()[:len(prefix)] != prefix {
		return
	}

	c := strings.Split(m.Message.ContentWithMentionsReplaced()[(len(prefix)+1):], " ")
	fmt.Println(c)

	for i := range c {
		c[i] = strings.TrimSpace(c[i])
	}

	if c[0] == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong!")
		return
	}

	if c[0] == "xkcd" {
		if xkcd, err := GetXkcd(c[1]); err != nil {
			fmt.Println(err)
		} else {
			s.ChannelMessageSend(m.ChannelID, xkcd.Img)
		}
	}
}
