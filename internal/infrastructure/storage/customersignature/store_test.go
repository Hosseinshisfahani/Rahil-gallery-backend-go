package customersignature

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestStoreSaveAndDelete(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	url, err := store.Save(id, "image/png", []byte("fake-png"))
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if url != "/static/customer-signatures/"+id.String()+".png" {
		t.Fatalf("unexpected url: %s", url)
	}

	path := filepath.Join(dir, id.String()+".png")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file missing: %v", err)
	}

	url2, err := store.Save(id, "image/jpeg", []byte("fake-jpg"))
	if err != nil {
		t.Fatalf("Save jpeg: %v", err)
	}
	if url2 != "/static/customer-signatures/"+id.String()+".jpg" {
		t.Fatalf("unexpected jpeg url: %s", url2)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("old png should be removed")
	}

	if err := store.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, id.String()+".jpg")); !os.IsNotExist(err) {
		t.Fatalf("jpeg should be removed")
	}
}

func TestStoreRejectsInvalidType(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	_, err := store.Save(uuid.New(), "text/plain", []byte("nope"))
	if err != ErrInvalidType {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}
