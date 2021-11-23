package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

func getOwnerAliases(filename string) (*Aliases, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Aliases{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getOwnersInfo(file string) (*OwnersInfo, error) {
	filename, _ := filepath.Abs(file)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &OwnersInfo{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getOwnerFiles(root string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if "OWNERS" == filepath.Base(path) && !strings.Contains(path, "vendor") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}
