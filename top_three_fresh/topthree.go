package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"strconv"
	"strings"

	"os"

	"sort"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	movies, err := getMovies()
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range movies {
		fmt.Println(m.Score, m.Name)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		content, err := json.Marshal(movies)
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

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

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

func getMovies() ([]*Movie, error) {
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
