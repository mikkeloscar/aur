package aur

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const aurURL = "https://aur.archlinux.org/rpc.php?"

type response struct {
	Error       string `json:"error"`
	Version     int    `json:"version"`
	Type        string `json:"type"`
	ResultCount int    `json:"resultcount"`
	Results     []Pkg  `json:"results"`
}

// Pkg holds package information
type Pkg struct {
	URL            string
	Description    string
	Version        string
	Name           string
	FirstSubmitted int
	License        []string
	ID             int
	PackageBaseID  int
	PackageBase    string
	OutOfDate      int
	LastModified   int
	Maintainer     string
	CategoryID     int
	URLPath        string
	NumVotes       int
	Conflicts      []string
	Depends        []string
	MakeDepends    []string
	OptDepends     []string
	Provides       []string
}

func get(values url.Values) ([]Pkg, error) {
	values.Set("v", "5")
	resp, err := http.Get(aurURL + values.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	result := new(response)
	err = dec.Decode(result)
	if err != nil {
		return nil, err
	}

	if len(result.Error) > 0 {
		return nil, fmt.Errorf(result.Error)
	}

	return result.Results, nil
}

// Search searches for packages
func Search(query string) ([]Pkg, error) {
	v := url.Values{}
	v.Set("type", "search")
	v.Set("arg", query)

	return get(v)
}

// MSearch searches for package by user
func MSearch(user string) ([]Pkg, error) {
	v := url.Values{}
	v.Set("type", "msearch")
	v.Set("arg", user)

	return get(v)
}

// Info shows info for one or multiple packages.
func Info(pkgs []string) ([]Pkg, error) {
	v := url.Values{}
	v.Set("type", "info")
	for _, arg := range pkgs {
		v.Add("arg[]", arg)
	}
	return get(v)
}
