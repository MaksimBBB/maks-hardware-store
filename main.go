package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Brand string `json:"brand"`
	Price int    `json:"price"`
}

func (i Item) Validate() error {
	if i.Name == "" || i.Brand == "" || i.Price <= 0 {
		return errors.New("invalid item data")
	}
	return nil
}

type Storage interface {
	Add(item Item) (Item, error)
	List() ([]Item, error)
	Get(id int) (Item, error)
	Delete(id int) error
}

var store Storage = NewFileStorage()

func addItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := item.Validate(); err != nil {
		http.Error(w, `{"error":"invalid item data"}`, http.StatusBadRequest)
		return
	}

	item, err := store.Add(item)
	if err != nil {
		http.Error(w, `{"error":"cannot create item"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func listItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	items, err := store.List()
	if err != nil {
		http.Error(w, `{"error":"cannot list items"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(items)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/item/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	item, err := store.Get(id)
	if err != nil {
		http.Error(w, `{"error":"item not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(item)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/delete/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := store.Delete(id); err != nil {
		http.Error(w, `{"error":"item not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

func main() {
	http.HandleFunc("/add", addItem)
	http.HandleFunc("/list", listItems)
	http.HandleFunc("/item/", getItem)
	http.HandleFunc("/delete/", deleteItem)

	http.ListenAndServe(":8080", nil)
}
