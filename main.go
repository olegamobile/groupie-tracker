package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
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

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" {
			http.Error(w, "URL not found", http.StatusNotFound)
			log.Println("URL not found:", r.URL.Path)
			return
		}

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
	})

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
