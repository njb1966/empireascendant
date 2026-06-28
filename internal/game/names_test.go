package game

import "testing"

func TestNormalizeName(t *testing.T) {
	name, err := NormalizeName(" New Terra ")
	if err != nil {
		t.Fatal(err)
	}
	if name.Display != "New Terra" {
		t.Fatalf("display = %q", name.Display)
	}
	if name.Key != "NEWTERRA" {
		t.Fatalf("key = %q", name.Key)
	}
}

func TestNormalizeNameRejectsLongName(t *testing.T) {
	_, err := NormalizeName("123456789012345678901")
	if err != ErrNameLong {
		t.Fatalf("err = %v", err)
	}
}
