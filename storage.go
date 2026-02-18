package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

//storage файлова логіка перенесена сюди

type FileStorage struct {
	dir string
}

func NewFileStorage() *FileStorage {
	os.MkdirAll("items", 0755)
	return &FileStorage{dir: "items"}
}

func (fs *FileStorage) Add(item Item) (Item, error) {
	files, _ := os.ReadDir(fs.dir)
	item.ID = len(files) + 1

	path := fmt.Sprintf("%s/item%d.json", fs.dir, item.ID)
	file, err := os.Create(path)
	if err != nil {
		return Item{}, err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(item); err != nil {
		return Item{}, err
	}

	return item, nil
}

func (fs *FileStorage) List() ([]Item, error) {
	files, err := os.ReadDir(fs.dir)
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, f := range files {
		file, _ := os.Open(fs.dir + "/" + f.Name())
		var item Item
		json.NewDecoder(file).Decode(&item)
		file.Close()
		items = append(items, item)
	}
	return items, nil
}

func (fs *FileStorage) Get(id int) (Item, error) {
	path := fmt.Sprintf("%s/item%d.json", fs.dir, id)
	file, err := os.Open(path)
	if err != nil {
		return Item{}, errors.New("not found")
	}
	defer file.Close()

	var item Item
	json.NewDecoder(file).Decode(&item)
	return item, nil
}

func (fs *FileStorage) Delete(id int) error {
	path := fmt.Sprintf("%s/item%d.json", fs.dir, id)
	return os.Remove(path)
}
