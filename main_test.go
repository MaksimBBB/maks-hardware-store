package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// мокування. Не залежаить від реального сховища
type MockStorage struct {
	AddFunc    func(Item) (Item, error)
	ListFunc   func() ([]Item, error)
	GetFunc    func(int) (Item, error)
	DeleteFunc func(int) error
}

func (m *MockStorage) Add(i Item) (Item, error) {
	return m.AddFunc(i)
}
func (m *MockStorage) List() ([]Item, error) {
	return m.ListFunc()
}
func (m *MockStorage) Get(id int) (Item, error) {
	return m.GetFunc(id)
}
func (m *MockStorage) Delete(id int) error {
	return m.DeleteFunc(id)
}

// табличний тест
func TestItemValidate(t *testing.T) {
	tests := []struct {
		name string
		item Item
		ok   bool
	}{
		{"valid", Item{Name: "Phone", Brand: "Apple", Price: 1000}, true},
		{"empty name", Item{Name: "", Brand: "Apple", Price: 1000}, false},
		{"empty brand", Item{Name: "Phone", Brand: "", Price: 1000}, false},
		{"zero price", Item{Name: "Phone", Brand: "Apple", Price: 0}, false},
	}

	for _, tt := range tests {
		err := tt.item.Validate()
		if tt.ok && err != nil {
			t.Errorf("%s: expected valid", tt.name)
		}
		if !tt.ok && err == nil {
			t.Errorf("%s: expected error", tt.name)
		}
	}
}

func TestAddItem(t *testing.T) {

	// invalid method
	req := httptest.NewRequest(http.MethodGet, "/add", nil)
	rr := httptest.NewRecorder()
	addItem(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Error("expected 405")
	}

	// invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/add", bytes.NewBuffer([]byte("{invalid")))
	rr = httptest.NewRecorder()
	addItem(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Error("expected 400 for invalid JSON")
	}

	// invalid data
	body, _ := json.Marshal(Item{Name: "", Brand: "A", Price: 10})
	req = httptest.NewRequest(http.MethodPost, "/add", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	addItem(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Error("expected 400 for invalid data")
	}

	// storage error
	store = &MockStorage{
		AddFunc: func(i Item) (Item, error) {
			return Item{}, errors.New("fail")
		},
	}
	body, _ = json.Marshal(Item{Name: "Phone", Brand: "A", Price: 10})
	req = httptest.NewRequest(http.MethodPost, "/add", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	addItem(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Error("expected 500")
	}

	// success
	store = &MockStorage{
		AddFunc: func(i Item) (Item, error) {
			i.ID = 1
			return i, nil
		},
	}
	req = httptest.NewRequest(http.MethodPost, "/add", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	addItem(rr, req)
	if rr.Code != http.StatusCreated {
		t.Error("expected 201")
	}
}

func TestListItems(t *testing.T) {

	// invalid method
	req := httptest.NewRequest(http.MethodPost, "/list", nil)
	rr := httptest.NewRecorder()
	listItems(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Error("expected 405")
	}

	// storage error
	store = &MockStorage{
		ListFunc: func() ([]Item, error) {
			return nil, errors.New("fail")
		},
	}
	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	rr = httptest.NewRecorder()
	listItems(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Error("expected 500")
	}

	// success
	store = &MockStorage{
		ListFunc: func() ([]Item, error) {
			return []Item{{ID: 1, Name: "Phone", Brand: "A", Price: 10}}, nil
		},
	}
	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	rr = httptest.NewRecorder()
	listItems(rr, req)
	if rr.Code != http.StatusOK {
		t.Error("expected 200")
	}
}

func TestGetItem(t *testing.T) {

	// invalid method
	req := httptest.NewRequest(http.MethodPost, "/item/1", nil)
	rr := httptest.NewRecorder()
	getItem(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Error("expected 405")
	}

	// invalid id
	req = httptest.NewRequest(http.MethodGet, "/item/abc", nil)
	rr = httptest.NewRecorder()
	getItem(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Error("expected 400")
	}

	// not found
	store = &MockStorage{
		GetFunc: func(id int) (Item, error) {
			return Item{}, errors.New("not found")
		},
	}
	req = httptest.NewRequest(http.MethodGet, "/item/1", nil)
	rr = httptest.NewRecorder()
	getItem(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Error("expected 404")
	}

	// success
	store = &MockStorage{
		GetFunc: func(id int) (Item, error) {
			return Item{ID: 1, Name: "Phone", Brand: "A", Price: 10}, nil
		},
	}
	req = httptest.NewRequest(http.MethodGet, "/item/1", nil)
	rr = httptest.NewRecorder()
	getItem(rr, req)
	if rr.Code != http.StatusOK {
		t.Error("expected 200")
	}
}

func TestDeleteItem(t *testing.T) {

	// invalid method
	req := httptest.NewRequest(http.MethodGet, "/delete/1", nil)
	rr := httptest.NewRecorder()
	deleteItem(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Error("expected 405")
	}

	// invalid id
	req = httptest.NewRequest(http.MethodDelete, "/delete/abc", nil)
	rr = httptest.NewRecorder()
	deleteItem(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Error("expected 400")
	}

	// not found
	store = &MockStorage{
		DeleteFunc: func(id int) error {
			return errors.New("not found")
		},
	}
	req = httptest.NewRequest(http.MethodDelete, "/delete/1", nil)
	rr = httptest.NewRecorder()
	deleteItem(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Error("expected 404")
	}

	// success
	store = &MockStorage{
		DeleteFunc: func(id int) error {
			return nil
		},
	}
	req = httptest.NewRequest(http.MethodDelete, "/delete/1", nil)
	rr = httptest.NewRecorder()
	deleteItem(rr, req)
	if rr.Code != http.StatusOK {
		t.Error("expected 200")
	}
}

func TestFileStorage_Real(t *testing.T) {
	os.RemoveAll("items")
	fs := NewFileStorage()

	item := Item{Name: "Phone", Brand: "Apple", Price: 100}

	// Add
	added, err := fs.Add(item)
	if err != nil {
		t.Fatal("add failed")
	}

	// Get
	got, err := fs.Get(added.ID)
	if err != nil || got.Name != "Phone" {
		t.Fatal("get failed")
	}

	// List
	list, err := fs.List()
	if err != nil || len(list) != 1 {
		t.Fatal("list failed")
	}

	// Delete
	err = fs.Delete(added.ID)
	if err != nil {
		t.Fatal("delete failed")
	}
}
