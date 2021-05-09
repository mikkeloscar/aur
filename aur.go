package aur

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

// AURURL is the base string from which the query is built.
var AURURL = "https://aur.archlinux.org/rpc.php?"

var (
	// ErrServiceUnavailable represents a error when AUR is unavailable.
	ErrServiceUnavailable = errors.New("AUR is unavailable at this moment")
)

// By specifies what to seach by in RPC searches.
type By int

const (
	Name By = iota + 1
	NameDesc
	Maintainer
	Depends
	MakeDepends
	OptDepends
	CheckDepends
	None
)

func (by By) String() string {
	switch by {
	case Name:
		return "name"
	case NameDesc:
		return "name-desc"
	case Maintainer:
		return "maintainer"
	case Depends:
		return "depends"
	case MakeDepends:
		return "makedepends"
	case OptDepends:
		return "optdepends"
	case CheckDepends:
		return "checkdepends"
	case None:
		return ""
	default:
		panic("invalid By")
	}
}

func get(values url.Values) ([]Pkg, error) {
	values.Set("v", "5")
	resp, err := http.Get(AURURL + values.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := getErrorByStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	dec := json.NewDecoder(resp.Body)
	result := new(response)
	err = dec.Decode(result)
	if err != nil {
		return nil, err
	}

	if len(result.Error) > 0 {
		return nil, errors.New(result.Error)
	}

	return result.Results, nil
}

func searchBy(query, by string) ([]Pkg, error) {
	v := url.Values{}
	v.Set("type", "search")
	v.Set("arg", query)

	if by != "" {
		v.Set("by", by)
	}

	return get(v)
}

func getErrorByStatusCode(code int) error {
	switch code {
	case http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	}
	return nil
}

// Search searches for packages using the RPC's default search by.
// This is the same as using SearchBy With NameDesc
func Search(query string) ([]Pkg, error) {
	return searchBy(query, "")
}

// SearchBy searches for packages with a specified  search by
func SearchBy(query string, by By) ([]Pkg, error) {
	return searchBy(query, by.String())
}

func Orphans() ([]Pkg, error) {
	return SearchBy("", Maintainer)
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
