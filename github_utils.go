package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func fetchPRCommentCount(user string) (int, error) {
	t := time.Now().AddDate(-1, 0, 0)
	url := "https://api.github.com/search/issues?q=" +
		"is%3Apr" +
		"+involves%3A" + user +
		"+is%3Amerged" +
		"+updated%3A>%3D" + t.Format("2006-01-02") +
		"+commenter%3A" + user +
		"+repo%3Akubernetes%2Fkubernetes" +
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
