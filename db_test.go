package codetainer

import (
	"os"
	"testing"
)

func deleteIfExists(file string) {
	if fileExists(file) {
		os.Remove(file)
	}
}

func TestDbCreation(t *testing.T) {

	deleteIfExists("/tmp/test.db")
	db, err := NewDatabase("/tmp/test.db")

	if err != nil {
		t.Fatal(err)
	}

	if db == nil {
		t.Fatal("no db returned")
	}

	if _, err := os.Stat("/tmp/test.db"); err != nil {
		t.Fatal("no db found")
	}
}
