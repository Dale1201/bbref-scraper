package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	s "strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Player struct {
	Name string
	IsActive bool
	isHoF bool
	Position string
	Height string
	Teams []string
	SeasonAverages []SeasonAverage
}

type SeasonAverage struct {
	Season string
	Age int
	Team string
	GamesPlayed int
	PtsPerGame float64
	RebPerGame float64
	AstPerGame float64
	StlPerGame float64
	BlkPerGame float64
	TOVPerGame float64
	FGP float64
	ThreePP float64
	FTPerGame float64
	MPG float64
	FTP float64
}

var LETTERS = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

func main() {
	// Instantiate default collector
	c := colly.NewCollector()
	playerC := colly.NewCollector()

	playerC.Limit(&colly.LimitRule{
		DomainGlob:  "*basketball-reference.com*",
		Parallelism: 2,
		Delay: 2 * time.Second,
	})

	players := map[string]Player{}

	// On every a element which has href attribute call callback
	c.OnHTML("tbody tr", func(e *colly.HTMLElement) {
		name := e.ChildText("th")
		playerProfileLink := e.ChildAttr("a", "href")
		position  := e.ChildText("td[data-stat='pos']")
		height := e.ChildText("td[data-stat='height']")

		var isActive bool
		if e.ChildText("strong") != "" {
			isActive = true
		} else {
			isActive = false
		}

		var isHoF bool
		if s.Contains(e.ChildText("th"), "*") {
			isHoF = true
			name = s.Replace(name, "*", "", -1)
		} else {
			isHoF = false
		}

		player := Player{name, isActive, isHoF, position, height, []string{}, []SeasonAverage{}}
		players[name] = player


		playerC.Visit("https://www.basketball-reference.com" + playerProfileLink)
		fmt.Println(player, playerProfileLink)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})


	playerC.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	playerC.OnHTML("body", func(e *colly.HTMLElement) {
		name := e.ChildText("h1")

		teams := []string{}
		teamsMap := map[string]bool{}
		table := e.DOM.Find("table#per_game tbody")

		seasonAverages := []SeasonAverage{}

		table.Children().Each(func(i int, s *goquery.Selection) {
			team := s.Find("td[data-stat='team_id']").Text()
			if team != "TOT" && !teamsMap[team] {
				teams = append(teams, team)
				teamsMap[team] = true
			}

			season := s.Find("th[data-stat='season']").Text()
			age, _ := strconv.Atoi(s.Find("td[data-stat='age']").Text())
			gamesPlayed, _ := strconv.Atoi(s.Find("td[data-stat='g']").Text())
			ptsPerGame, _ := strconv.ParseFloat(s.Find("td[data-stat='pts_per_g']").Text(), 64)
			rebPerGame, _ := strconv.ParseFloat(s.Find("td[data-stat='trb_per_g']").Text(), 64)
			astPerGame, _ := strconv.ParseFloat(s.Find("td[data-stat='ast_per_g']").Text(), 64)
			stlPerGame, _ :=  strconv.ParseFloat(s.Find("td[data-stat='stl_per_g']").Text(), 64)
			blkPerGame, _ :=  strconv.ParseFloat(s.Find("td[data-stat='blk_per_g']").Text(), 64)
			tovPerGame, _ :=  strconv.ParseFloat(s.Find("td[data-stat='tov_per_g']").Text(), 64)
			fgp, _ :=  strconv.ParseFloat(s.Find("td[data-stat='fg_pct']").Text(), 64)
			threePP, _ :=  strconv.ParseFloat(s.Find("td[data-stat='fg3_pct']").Text(), 64)
			ftPerGame, _ :=  strconv.ParseFloat(s.Find("td[data-stat='ft_per_g']").Text(), 64)
			mpg, _ :=  strconv.ParseFloat(s.Find("td[data-stat='mp_per_g']").Text(), 64)
			ftP, _ :=  strconv.ParseFloat(s.Find("td[data-stat='ft_pct']").Text(), 64)


			seasonAverage := SeasonAverage{season, age, team, gamesPlayed, ptsPerGame, rebPerGame, astPerGame, stlPerGame, blkPerGame, tovPerGame, fgp, threePP, ftPerGame, mpg, ftP}
			seasonAverages = append(seasonAverages, seasonAverage)

		})

		player := players[name]
		player.Teams = teams
		player.SeasonAverages = seasonAverages
		players[name] = player

	})

	for _, letter := range LETTERS {
		c.Visit("https://www.basketball-reference.com/players/" + letter + "/")
	}

	// c.Visit("https://www.basketball-reference.com/players/a/")

	// Write players to json file
	jsonPlayers, err := json.Marshal(players)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("players.json", jsonPlayers, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done")
}