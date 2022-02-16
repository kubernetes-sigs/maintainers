/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

func GetOwnerAliases(filename string) (*Aliases, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Aliases{}
	err = yaml.UnmarshalStrict(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func GetSigsYaml(filename string) (*Context, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Context{}
	err = yaml.UnmarshalStrict(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func GetOwnersInfo(file string) (*OwnersInfo, error) {
	filename, _ := filepath.Abs(file)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config, err := GetOwnersInfoFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func GetOwnersInfoFromBytes(bytes []byte) (*OwnersInfo, error) {
	config := &OwnersInfo{}
	err := yaml.UnmarshalStrict(bytes, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func GetOwnersAliasesFile(root string) (string, error) {
	var err error
	aliasPath, _ := filepath.Abs(filepath.Join(root, "OWNERS_ALIASES"))
	if _, err = os.Stat(aliasPath); err == nil {
		return aliasPath, nil
	}
	return "", err
}

func GetOwnerFiles(root string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == "OWNERS" && !strings.Contains(path, "vendor") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func GetSigsYamlFile(root string) (string, error) {
	var err error
	path, _ := filepath.Abs(filepath.Join(root, "sigs.yaml"))
	if _, err = os.Stat(path); err == nil {
		return path, nil
	}
	return "", err
}
