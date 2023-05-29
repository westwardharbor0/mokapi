package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Definition struct {
	Endpoint           string          `json:"endpoint"`
	Method             string          `json:"method"`
	ResponseStatusCode int             `json:"response_status_code"`
	ResponsePayload    json.RawMessage `json:"response_payload"`

	LastState os.FileInfo `json:"-"`
	Path      string      `json:"-"`
}

func (d *Definition) Changed() (bool, error) {
	stat, err := os.Stat(d.Path)
	if err != nil {
		return false, err
	}
	changed := stat.Size() != d.LastState.Size() || stat.ModTime() != d.LastState.ModTime()
	d.LastState = stat
	return changed, nil
}

type Definitions struct {
	Path      string
	Endpoints map[string]*Definition
}

func (d *Definitions) CheckPresence() error {
	info, err := os.Stat(d.Path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("path to fileDefinitions is a file")
	}
	return nil
}

func (d *Definitions) Load() error {
	if err := d.CheckPresence(); err != nil {
		return err
	}

	d.Endpoints = make(map[string]*Definition)
	folderItems, err := os.ReadDir(d.Path)
	if err != nil {
		return err
	}

	for _, folderItem := range folderItems {
		itemPath := filepath.Join(d.Path, folderItem.Name())
		definition, err := loadDefinitionFromFile(itemPath)
		if err != nil {
			return err
		}
		if err := d.Add(definition); err != nil {
			return err
		}
	}
	return nil
}

func (d *Definitions) Add(definition *Definition) error {
	key := fmt.Sprintf("%s:%s", definition.Method, definition.Endpoint)
	if _, exists := d.Endpoints[key]; exists {
		return fmt.Errorf("the endpoint %s is already used and set up", key)
	}
	d.Endpoints[key] = definition
	return nil
}

func loadDefinitionFromFile(path string) (*Definition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var fileContent []byte
	fileContent, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var definition Definition
	err = json.Unmarshal(fileContent, &definition)
	if err != nil {
		return nil, err
	}

	definition.Method = strings.ToUpper(definition.Method)
	stats, err := file.Stat()
	definition.LastState = stats
	definition.Path = path
	if err != nil {
		return nil, err
	}

	err = file.Close()
	return &definition, err
}
