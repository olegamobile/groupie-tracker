package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
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

// getting data from the given API
func FetchGroups() ([]Group, error) {
	endpointURL := "https://groupietrackers.herokuapp.com/api/artists"
	resp, err := http.Get(endpointURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []Group
	// decoding JSON data into "groups" using pointer to it
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func fetchGroup(id int) (Group, error) {
	endpointURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/artists/%d", id)
	resp, err := http.Get(endpointURL)
	if err != nil {
		return Group{}, err
	}
	defer resp.Body.Close()

	var group Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return Group{}, err
	}
	return group, nil
}

func fetchRelation(id int) (Relation, error) {
	endpointURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", id)
	resp, err := http.Get(endpointURL)
	if err != nil {
		return Relation{}, err
	}
	defer resp.Body.Close()

	var relation Relation
	if err := json.NewDecoder(resp.Body).Decode(&relation); err != nil {
		return Relation{}, err
	}
	return relation, nil
}

func main() {
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.URL.Query().Get("id") == "" {
			groups, err := FetchGroups()
			if err != nil {
				http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
				log.Println("Error fetching data from API:", err)
				return
			}

			tmpl, err := template.ParseFiles("templates/index.html")

			if err != nil {
				http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
				log.Println("Error parsing template:", err)
				return
			}
			tmpl.Execute(w, groups)
		} else if r.URL.Query().Get("id") != "" {
			idStr := r.URL.Query().Get("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			group, err := fetchGroup(id)
			if err != nil {
				http.Error(w, "Error fetching group data", http.StatusInternalServerError)
				return
			}

			relation, err := fetchRelation(id)
			if err != nil {
				http.Error(w, "Error fetching relation data", http.StatusInternalServerError)
				return
			}

			// Render the template
			tmpl, err := template.ParseFiles("templates/details.html")
			if err != nil {
				http.Error(w, "Error loading template", http.StatusInternalServerError)
				return
			}
			data := struct {
				Artist   Group
				Relation Relation
			}{
				Artist:   group,
				Relation: relation,
			}
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
				return
			}

		} else {
			if r.URL.Path != "/" {
				http.Error(w, "URL not found", http.StatusNotFound)
				log.Println("URL not found:", r.URL.Path)
				return
			}
		}
	})

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
