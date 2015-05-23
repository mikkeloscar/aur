package aur

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const aurURL = "https://aur.archlinux.org/rpc.php?"

type infoResp struct {
	Version     int    `json:"version"`
	Type        string `json:"type"`
	ResultCount int    `json:"resultcount"`
	Results     Pkg    `json:"results"`
}

type multiinfoResp struct {
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
	License        string
	ID             int
	PackageBaseID  int
	PackageBase    string
	OutOfDate      int
	LastModified   int
	Maintainer     string
	CategoryID     int
	URLPath        string
	NumVotes       int
}

func get(values url.Values) ([]Pkg, error) {
	resp, err := http.Get(aurURL + values.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	result := new(multiinfoResp)
	err = dec.Decode(result)
	if err != nil {
		return nil, err
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

// Info shows info of package pkg
func Info(pkg string) (*Pkg, error) {
	v := url.Values{}
	v.Set("type", "info")
	v.Set("arg", pkg)

	resp, err := http.Get(aurURL + v.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	result := new(infoResp)
	err = dec.Decode(result)
	if err != nil {
		return nil, err
	}

	return &result.Results, nil
}

// Multiinfo shows info for multiple packages
func Multiinfo(pkgs []string) ([]Pkg, error) {
	v := url.Values{}
	v.Set("type", "multiinfo")
	for _, arg := range pkgs {
		v.Add("arg[]", arg)
	}
	return get(v)
}
