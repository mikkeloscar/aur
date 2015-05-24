package aur

import "testing"

// TestSearch test searching for packages by name and/or description
func TestSearch(t *testing.T) {
	rs, err := Search("linux")
	if err != nil {
		t.Error(err)
	}

	if len(rs) < 100 {
		t.Errorf("Expected more than 100 packages, got '%d'", len(rs))
	}

	rs, err = Search("li")
	if err.Error() != "Too many package results." {
		t.Errorf("Expected error 'Too many package results.', got '%s'", err.Error())
	}

	if len(rs) > 0 {
		t.Errorf("Expected no results, got '%d'", len(rs))
	}
}

// TestMSearch test searching for packages by maintainer
func TestMSearch(t *testing.T) {
	rs, err := MSearch("moscar")
	if err != nil {
		t.Error(err)
	}

	if len(rs) < 3 {
		t.Errorf("Expected more than 3 packages, got '%d'", len(rs))
	}
}

// TestInfo test getting info for single package
func TestInfo(t *testing.T) {
	rs, err := Info("rofi")
	if err != nil {
		t.Error(err)
	}

	if rs.Name != "rofi" {
		t.Errorf("Expected package name 'rofi', got '%s'", rs.Name)
	}
}

// TestMultiinfo test getting info for multiple packages
func TestMultiinfo(t *testing.T) {
	rs, err := Multiinfo([]string{"neovim-git", "rofi"})
	if err != nil {
		t.Error(err)
	}

	if len(rs) != 2 {
		t.Errorf("Expected two packages, got %d", len(rs))
	}
}
