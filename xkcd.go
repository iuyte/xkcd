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
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
