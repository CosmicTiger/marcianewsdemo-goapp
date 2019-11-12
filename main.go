package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Struct for working with the data from News API

type Source struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}

type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Results    Results
}

// DEFINITION OF THE HANDLER Variables
var tpl = template.Must(template.ParseFiles("index.html"))

var apiKey *string

// This function handles the index page
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("<h1>Hello World!</h1>"))
	tpl.Execute(w, nil)
}

// This function handles the search result page
func searchHandler(w http.ResponseWriter, r *http.Request) {

	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	// start of the function

	search := &Search{}
	search.SearchKey = searchKey

	next, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}

	search.NextPage = next
	pageSize := 20

	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	// Adding this if block
	if ok := !search.IsLastPage(); ok {
		search.NextPage++
	}

	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)

	resp, err := http.Get(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	err = tpl.Execute(w, search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Replacing this functions
	/*
		u, err := url.Parse(r.URL.String())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
			return
		}

		params := u.Query()
		searchKey := params.Get("q")
		page := params.Get("page")
		if page == "" {
			page = "1"
		}
		fmt.Println("Search Query is: ", searchKey)
		fmt.Println("Results page is: ", page) */
}

// Format the published date
func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

// Pagination on Go
func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}

	return s.NextPage - 1
}

func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

// main.go program

func main() {

	// adding the apiKey for the app

	apiKey = flag.String("apikey", "", "Newsap.org access key")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("apiKey must be set")
	}

	// Defining the port for the execution of the application

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	// mux is a variable for HTTP request multiplexer

	mux := http.NewServeMux()

	// Adding the handler for static files
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Routes of the application
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux)
}

// API_KEY = 9a11f05a76df4bf78d36729abea30728
// Command to execute this application
// go run main.go -apikey=9a11f05a76df4bf78d36729abea30728
