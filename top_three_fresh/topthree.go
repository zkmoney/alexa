package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"strconv"
	"strings"

	"os"

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
		w.WriteHeader(200)
		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&movies); err != nil {
			log.Println("error writing top three:", err)
		}
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

func getMovies() ([]*Movie, error) {
	url := "https://www.rottentomatoes.com"
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	var movies []*Movie
	doc.Find("#Top-Box-Office tr").Each(func(i int, s *goquery.Selection) {
		movies = append(movies, &Movie{
			Name:  s.Find(".middle_col a").Text(),
			Score: scoreToInt(s.Find(".left_col .tMeterScore").Text()),
		})
	})

	return movies, nil
}

func scoreToInt(score string) int {
	i, _ := strconv.Atoi(strings.Replace(score, "%", "", -1))
	return i
}
