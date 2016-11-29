package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"strconv"
	"strings"

	"os"

	"sort"

	"time"

	"github.com/PuerkitoBio/goquery"
)

type Movie struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type Movies []*Movie

type ByScore struct {
	Movies
}

func (m Movies) Len() int {
	return len(m)
}

func (m Movies) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (b ByScore) Less(i, j int) bool {
	return b.Movies[i].Score < b.Movies[j].Score
}

func getMovies() (Movies, error) {
	url := "https://www.rottentomatoes.com"
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	var movies Movies
	doc.Find("#Top-Box-Office tr").Each(func(i int, s *goquery.Selection) {
		movies = append(movies, &Movie{
			Name:  s.Find(".middle_col a").Text(),
			Score: scoreToInt(s.Find(".left_col .tMeterScore").Text()),
		})
	})

	sort.Sort(sort.Reverse(ByScore{movies}))

	return movies, nil
}

func scoreToInt(score string) int {
	i, _ := strconv.Atoi(strings.Replace(score, "%", "", -1))
	return i
}

type MovieList struct {
	sync.Mutex
	movies Movies
}

func (ml *MovieList) update() error {
	ml.Lock()
	defer ml.Unlock()
	movies, err := getMovies()
	if err != nil {
		return err
	}
	ml.movies = movies
	return nil
}

func (ml *MovieList) Init() error {
	return ml.update()
}

func (ml *MovieList) Run() {
	for {
		if err := ml.update(); err != nil {
			log.Println("error updating movies: ", err)
		} else {
			log.Println("movies updated", ml.Movies())
		}
		time.Sleep(2 * time.Minute)
	}
}

func (ml *MovieList) Movies() Movies {
	ml.Lock()
	movies := ml.movies
	ml.Unlock()
	return movies
}

var movieList = &MovieList{}

func main() {
	if err := movieList.Init(); err != nil {
		log.Fatal(err)
	}

	// Start the movie updater
	go movieList.Run()

	if err := startServer(); err != nil {
		log.Fatal(err)
	}
}

func startServer() error {
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		content, err := json.Marshal(movieList.Movies())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return http.ListenAndServe(":"+port, nil)
}
