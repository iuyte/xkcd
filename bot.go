package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	prefix string = ";"
	token  string
	latest XKCD
)

func main() {
	var err error
	latest, err = GetLatest()
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			<-t.C
			rand.Seed(time.Now().Unix())
		}
	}()

	token = Token()
	if token == "" {
		fmt.Println("No token provided. Please place a token followed by a ';' at token.txt")
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
	b, err := ioutil.ReadFile("token.txt")
	if err != nil {
		fmt.Println(err)
	}
	c := string(b)
	return strings.Split(c, ";")[0]
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, prefix+"help")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Message.Content) < 1 {
		return
	}
	if m.Message.ContentWithMentionsReplaced()[:len(prefix)] != prefix {
		return
	}

	c := strings.Split(strings.TrimSpace(strings.TrimPrefix(m.Message.Content, prefix)), " ")
	c[0] = strings.ToLower(c[0])

	if c[0] == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong!")
		return
	}

	if c[0] == "xkcd" {
		if len(c) < 2 {
			c = append(c, "")
		}
		var (
			xkcd XKCD
			err  error
			r    *regexp.Regexp
		)

		r, err = regexp.Compile("^[0-9]+$")
		if err != nil {
			fmt.Println(err)
			return
		}

		if r.MatchString(c[1]) {
			xkcd, err = GetXkcdNum(c[1])
		} else {
			xkcd, err = GetXkcdTitle(strings.Join(c[1:], " "))
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		e := &discordgo.MessageEmbed{
			Title:       "xkcd #" + strconv.Itoa(xkcd.Num) + ": " + xkcd.Title,
			Description: xkcd.Alt,
			URL:         "https://xkcd.com/" + strconv.Itoa(xkcd.Num),
			Color:       7506394,
			Type:        "rich",
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		_, err = s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, xkcd.Img)
		}
		return
	}

	if c[0] == "latest" {
		xkcd := latest
		e := &discordgo.MessageEmbed{
			Title:       "xkcd #" + strconv.Itoa(xkcd.Num) + ": " + xkcd.Title,
			Description: xkcd.Alt,
			URL:         "https://xkcd.com/" + strconv.Itoa(xkcd.Num),
			Color:       7506394,
			Type:        "rich",
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		_, err := s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, xkcd.Img)
		}
		return
	}

	if c[0] == "random" {
		xkcd, err := GetXkcdNum(strconv.Itoa(rand.Intn(latest.Num + 1)))
		if err != nil {
			fmt.Println(err)
		}
		e := &discordgo.MessageEmbed{
			Title:       "xkcd #" + strconv.Itoa(xkcd.Num) + ": " + xkcd.Title,
			Description: xkcd.Alt,
			URL:         "https://xkcd.com/" + strconv.Itoa(xkcd.Num),
			Color:       7506394,
			Type:        "rich",
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		_, err = s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, xkcd.Img)
		}
		return
	}

	if c[0] == "help" {
		e := &discordgo.MessageEmbed{
			Title:       "Help",
			Description: "How to use this here xkcd bot",
			URL:         "https://xkcd.com/",
			Color:       7506394,
			Type:        "rich",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:  "Help",
					Value: "Display this message",
				},
				&discordgo.MessageEmbedField{
					Name:  "xkcd <comic number, name, regex or whatever>",
					Value: "Get the designated comic",
				},
				&discordgo.MessageEmbedField{
					Name:  "latest",
					Value: "The latest and greatest xkcd",
				},
				&discordgo.MessageEmbedField{
					Name:  "random",
					Value: "A random comic",
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		_, err := s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
		}
	}
}
