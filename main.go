package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Brand string `json:"brand"`
	Price int    `json:"price"`
}

// post - /add
func addItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if item.Name == "" || item.Brand == "" || item.Price <= 0 {
		http.Error(w, `{"error":"invalid item data"}`, http.StatusBadRequest)
		return
	}

	files, _ := os.ReadDir("items")
	item.ID = len(files) + 1

	filePath := fmt.Sprintf("items/item%d.json", item.ID)
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, `{"error":"cannot create file"}`, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	json.NewEncoder(file).Encode(item)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// get - /list
func listItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	files, _ := os.ReadDir("items")
	items := []Item{}

	for _, f := range files {
		file, _ := os.Open("items/" + f.Name())
		var item Item
		json.NewDecoder(file).Decode(&item)
		file.Close()
		items = append(items, item)
	}

	json.NewEncoder(w).Encode(items)
}

// get- /item/{id}
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

	filePath := fmt.Sprintf("items/item%d.json", id)
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, `{"error":"item not found"}`, http.StatusNotFound)
		return
	}
	defer file.Close()

	var item Item
	json.NewDecoder(file).Decode(&item)
	json.NewEncoder(w).Encode(item)
}

// delete - /delete/{id}
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

	filePath := fmt.Sprintf("items/item%d.json", id)
	err = os.Remove(filePath)
	if err != nil {
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
