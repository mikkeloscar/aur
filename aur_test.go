package aur

import "testing"

func expectPackages(t *testing.T, n int, rs []Pkg, err error) {
	if err != nil {
		t.Error(err)
	}

	if len(rs) < n {
		t.Errorf("Expected more than %d packages, got '%d'", n, len(rs))
	}
}


func expectTooMany(t *testing.T, rs []Pkg, err error) {
	if err.Error() != "Too many package results." {
		t.Errorf("Expected error 'Too many package results.', got '%s'", err.Error())
	}

	if len(rs) > 0 {
		t.Errorf("Expected no results, got '%d'", len(rs))
	}
}

// TestInfo test getting info for multiple packages
func TestInfo(t *testing.T) {
	rs, err := Info([]string{"neovim-git", "linux-mainline"})
	if err != nil {
		t.Error(err)
	}

	if len(rs) != 2 {
		t.Errorf("Expected two packages, got %d", len(rs))
	}
}

// TestSearch test searching for packages by the AUR's default by field
func TestSearch(t *testing.T) {
	rs, err := Search("linux")
	expectPackages(t, 100, rs, err)

	rs, err = Search("li")
	expectTooMany(t, rs, err)
}

// TestSearchByName test searching for packages by package name
func TestSearchByName(t *testing.T) {
	rs, err := SearchByNameDesc("linux")
	expectPackages(t, 100, rs, err)
}

// TestSearchByNameDesc test searching for packages package name and desc.
func TestSearchByNameDesc(t *testing.T) {
	rs, err := SearchByNameDesc("linux")
	expectPackages(t, 100, rs, err)
}

// TestSearchByMaintainer test searching for packages by maintainer
func TestSearchByMaintainer(t *testing.T) {
	rs, err := SearchByMaintainer("moscar")
	expectPackages(t, 3, rs, err)
}

// TestOrphans test searching for orphans
func TestOrphans(t *testing.T) {
	rs, err := Orphans()
	expectPackages(t, 500, rs, err)
}

// TestSearchByDepends test searching for packages by depends
func TestSearchByDepends(t *testing.T) {
	rs, err := SearchByDepends("python")
	expectPackages(t, 100, rs, err)
}

// TestSearchByMakeDepends test searching for packages by makedepends
func TestSearchByMakeDepends(t *testing.T) {
	rs, err := SearchByMakeDepends("python")
	expectPackages(t, 100, rs, err)
}

// TestSearchByOptDepends test searching for packages by optdepends
func TestSearchByOptDepends(t *testing.T) {
	rs, err := SearchByOptDepends("python")
	expectPackages(t, 100, rs, err)
}

// TestSearchByCheckDepends test searching for packages by checkdepends
func TestSearchByCheckDepends(t *testing.T) {
	rs, err := SearchByCheckDepends("python")
	expectPackages(t, 10, rs, err)
}
