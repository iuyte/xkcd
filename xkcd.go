package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
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

func GetLatest() (xkcd XKCD, e error) {
	var (
		r    *http.Response
		data []byte
	)

	r, e = http.Get("https://xkcd.com/info.0.json")
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

func GetXkcdNum(num string) (xkcd XKCD, e error) {
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

type Rating struct {
	Xkcd  XKCD
	Score int
}

func ratingSort(a []Rating) []Rating {
	if len(a) < 2 {
		return a
	}

	left, right := 0, len(a)-1

	pivotIndex := rand.Int() % len(a)

	a[pivotIndex], a[right] = a[right], a[pivotIndex]

	for i := range a {
		if a[i].Score < a[right].Score {
			a[i], a[left] = a[left], a[i]
			left++
		}
	}

	a[left], a[right] = a[right], a[left]

	ratingSort(a[:left])
	ratingSort(a[left+1:])

	return a
}

func GetXkcdTitle(title string) (xkcd XKCD, e error) {
	title = strings.ToLower(title)
	var (
		data    []byte
		ratings []Rating = make([]Rating, 10)
		r       *http.Response
		t       *regexp.Regexp
		last    bool = false
	)

	t, e = regexp.Compile(title)
	if e != nil {
		return
	}

	for i := 1; !last; i++ {
		r, e = http.Get("https://xkcd.com/" + strconv.Itoa(i) + "/info.0.json")
		if e != nil {
			break
		}

		data, e = func() ([]byte, error) {
			defer r.Body.Close()
			return ioutil.ReadAll(r.Body)
		}()
		if e != nil {
			return
		}

		e = json.Unmarshal(data, &xkcd)
		if e != nil {
			e = nil
			if i == 404 {
				continue
			}
			break
		}

		var rating Rating
		rating.Xkcd = xkcd
		rating.Score = len(t.FindAll(data, -1))
		if strings.Contains(strings.ToLower(xkcd.Title), title) {
			rating.Score += 2
		}
		if rating.Score > 1 {
			ratings = append(ratings, rating)
		}
	}
	xkcd = ratingSort(ratings)[len(ratings)-1].Xkcd

	return
}

const prefix string = ";"

var (
	token string
)

func main() {
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
	b, err := ioutil.ReadFile("/token.txt")
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

	if c[0] == "latest" {
		xkcd, err := GetLatest()
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
}
