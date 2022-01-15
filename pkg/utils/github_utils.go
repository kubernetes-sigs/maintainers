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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func FetchPRCommentCount(user, repository string) (int, error) {
	t := time.Now().AddDate(-1, 0, 0)
	url := "https://api.github.com/search/issues?q=" +
		"is%3Apr" +
		"+involves%3A" + user +
		"+is%3Amerged" +
		"+updated%3A>%3D" + t.Format("2006-01-02") +
		"+commenter%3A" + user +
		"+repo%3A" + url.QueryEscape(repository) +
		"+user%3A" + user

	spaceClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return -1, err
	}

	if token := os.Getenv("GITHUB_TOKEN"); len(token) != 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	res, err := spaceClient.Do(req)
	if err != nil {
		return -1, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode == http.StatusForbidden {
		return -1, nil
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return -1, readErr
	}

	var result map[string]interface{}
	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		return -1, jsonErr
	}

	return strconv.Atoi(fmt.Sprintf("%v", result["total_count"]))
}

func GetKubernetesOwnersFiles() (*[]string, error) {
	resp, err := http.Get("https://api.github.com/repos/kubernetes/kubernetes/git/trees/master?recursive=1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if dec == nil {
		panic("Failed to start decoding JSON data")
	}

	type Content struct {
		Files []struct {
			Path string
		} `json:"tree"`
	}

	c := &Content{}
	err = dec.Decode(&c)
	if err != nil {
		return nil, err
	}

	directories := make([]string, len(c.Files))
	for _, directory := range c.Files {
		if len(directory.Path) > 0 &&
			strings.Index(directory.Path, "/OWNERS") != -1 &&
			strings.Index(directory.Path, "vendor/") != 0 {
			directories = append(directories, directory.Path)
		}
	}
	return &directories, nil
}
