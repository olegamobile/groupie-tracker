package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// using struct tags because we should create exportable var names (starting from capital letters) to use them in templates
type Group struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Image      string   `json:"image"`
	Members    []string `json:"members"`
	Creation   int      `json:"creationDate"`
	FirstAlbum string   `json:"firstAlbum"`
}

type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type Location struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

type Dates struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

type Details struct {
	Artist       Group
	Relation     Relation
	Locations    string
	FirstConcert string
	LastConcert  string
}

const (
	ArtistURL   = "https://groupietrackers.herokuapp.com/api/artists"
	LocationURL = "https://groupietrackers.herokuapp.com/api/locations"
	DateURL     = "https://groupietrackers.herokuapp.com/api/dates"
	RelationURL = "https://groupietrackers.herokuapp.com/api/relation"
)

func fetchData(endpointURL, id string) (*http.Response, error) {
	if id != "" {
		endpointURL += "/" + id
	}
	resp, err := http.Get(endpointURL)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func FetchGroups() ([]Group, error) {
	resp, err := fetchData(ArtistURL, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []Group

	// decoding JSON data into "groups" using pointer to it
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func FetchGroup(id string) (Group, error) {
	resp, err := fetchData(ArtistURL, id)
	if err != nil {
		return Group{}, err
	}
	defer resp.Body.Close()

	var group Group

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&group); err != nil {
		return Group{}, err
	}

	return group, nil
}

func fetchRelation(id string) (Relation, error) {
	resp, err := fetchData(RelationURL, id)
	if err != nil {
		return Relation{}, err
	}
	defer resp.Body.Close()

	var relation Relation
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&relation); err != nil {
		return Relation{}, err
	}

	return relation, nil
}

func fetchLocations(id string) (string, error) {
	resp, err := fetchData(LocationURL, id)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var locations Location
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&locations); err != nil {
		return "", err
	}

	var countriesList []string
	countriesMap := make(map[string]bool)
	for _, location := range locations.Locations {
		_, country := getLocation(location)
		if !countriesMap[country] {
			countriesList = append(countriesList, country)
			countriesMap[country] = true
		}
	}
	return strings.Join(countriesList, ", "), nil
}

func getLocation(rawLocation string) (string, string) {
	var result [2]string
	elements := strings.Split(rawLocation, "-")

	for i, element := range elements {
		if i > 1 {
			break
		}
		parts := strings.Split(element, "_")
		if element == "usa" || element == "uk" {
			result[i] = strings.ToUpper(element)
			continue
		}
		var tempResult []string
		for _, part := range parts {
			tempResult = append(tempResult, strings.ToUpper(string(part[0]))+part[1:])
		}
		result[i] = strings.Join(tempResult, " ")
	}
	return result[0], result[1]
}

func fetchDates(id string) ([]string, error) {
	resp, err := fetchData(DateURL, id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dates Dates
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&dates); err != nil {
		return nil, err
	}
	result := []string{}
	for _, element := range dates.Dates {
		result = append(result, strings.ReplaceAll(element, "*", ""))
	}
	return result, nil
}

func getFirstLastDate(dates []string) (string, string, error) {
	inputFormat := "02-01-2006"
	outputFormat := "02 January 2006"
	if len(dates) == 0 {
		return "", "", fmt.Errorf("empty slice of dates")
	}
	minDate, err := time.Parse(inputFormat, dates[0])
	if err != nil {
		return "", "", fmt.Errorf("error parsing date %q: %v", dates[0], err)
	}
	maxDate := minDate
	for _, dateStr := range dates[1:] {
		parsedDate, err := time.Parse(inputFormat, dateStr)
		if err != nil {
			return "", "", fmt.Errorf("error parsing date %q: %v", dateStr, err)
		}
		if parsedDate.Before(minDate) {
			minDate = parsedDate
		}
		if parsedDate.After(maxDate) {
			maxDate = parsedDate
		}
	}
	return minDate.Format(outputFormat), maxDate.Format(outputFormat), nil
}

func generateMainPage(w http.ResponseWriter) {
	groups, err := FetchGroups()
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error fetching data from API:", err)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")

	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error parsing template:", err)
		return
	}
	tmpl.Execute(w, groups)
}

func generateDetailsPage(w http.ResponseWriter, id string) {
	group, err := FetchGroup(id)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error fetching group data:", err)
		return
	}

	relation, err := fetchRelation(id)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error fetching relations data:", err)
		return
	}

	countriesList, err := fetchLocations(id)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error fetching locations data:", err)
		return
	}

	tmpl, err := template.ParseFiles("templates/details.html")
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error parsing template:", err)
		return
	}

	dates, err := fetchDates(id)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error fetching dates from API:", err)
		return
	}
	var details Details
	details.Artist = group
	details.Relation = relation
	details.Locations = countriesList
	details.FirstConcert, details.LastConcert, err = getFirstLastDate(dates)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error calculation first and last dates:", err)
		return
	}

	if err := tmpl.Execute(w, details); err != nil {
		errorHandler(w, http.StatusInternalServerError, "Error executing template:", err)
		return
	}
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		if r.URL.Query().Get("id") == "" {
			generateMainPage(w)
		} else {
			id := r.URL.Query().Get("id")
			generateDetailsPage(w, id)
		}
	} else {
		errorHandler(w, http.StatusNotFound, "URL not found: "+r.URL.Path, nil)
		return
	}
}

func errorHandler(w http.ResponseWriter, statusCode int, errorMessage string, errorDetails error) {
	log.Println(errorMessage, errorDetails)
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	tmpl.ExecuteTemplate(w, "error.html", http.StatusText(statusCode))
}

func main() {
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", mainPageHandler)
	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
