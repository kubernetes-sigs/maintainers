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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func GetContributionsForAYear(repository string, period string) ([]Contribution, error) {
	originalRepository := repository
	postBody := `{
	"queries": [{
		"refId": "A",
		"datasourceId": 1,
		"rawSql": "select sub.name as name, sub.value from (select row_number() over (order by sum(value) desc) as \"Rank\", split_part(name, '$$$', 1) as name, sum(value) as value from shdev_repos where series = 'hdev_contributionsallall' and period = '%s' group by split_part(name, '$$$', 1)) sub",
		"format": "table"
	}]
}`
	repository = strings.Replace(repository, "/", "", -1)
	repository = strings.Replace(repository, "-", "", -1)
	repository = strings.Replace(repository, ".", "", -1)
	postBody = strings.Replace(
		fmt.Sprintf(postBody, period),
		"hdev_contributionsallall",
		fmt.Sprintf("hdev_contributions%sall", repository),
		-1)

	requestBody := bytes.NewBuffer([]byte(postBody))
	resp, err := http.Post("https://k8s.devstats.cncf.io/api/ds/query", "application/json", requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("devstats returned HTTP %d for repository %q period %q", resp.StatusCode, originalRepository, period)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	contribs, err := parseContributions(body)
	if err != nil {
		return nil, fmt.Errorf("unable to read devstats contributions for repository %q period %q: %w", originalRepository, period, err)
	}
	return contribs, nil
}

func parseContributions(body []byte) ([]Contribution, error) {
	var parsed map[string]map[string]map[string][]Frames
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("unable to parse json from devstats: %w", err)
	}

	results, ok := parsed["results"]
	if !ok {
		return nil, fmt.Errorf("response missing results")
	}
	queryResult, ok := results["A"]
	if !ok {
		return nil, fmt.Errorf("response missing query result A")
	}
	frames, ok := queryResult["frames"]
	if !ok || len(frames) == 0 {
		return nil, fmt.Errorf("response contained no frames")
	}

	items := frames[0].Data.Items
	if len(items) < 2 {
		return nil, fmt.Errorf("response contained %d data columns, want at least 2", len(items))
	}

	names := items[0]
	values := items[1]
	if len(names) == 0 {
		return nil, fmt.Errorf("response contained no contribution rows")
	}
	if len(names) != len(values) {
		return nil, fmt.Errorf("response contained mismatched data column lengths: %d names, %d values", len(names), len(values))
	}

	contribs := make([]Contribution, 0, len(names))
	for i := range names {
		name, ok := names[i].(string)
		if !ok {
			return nil, fmt.Errorf("response row %d has non-string contributor name %T", i, names[i])
		}
		value, ok := values[i].(float64)
		if !ok {
			return nil, fmt.Errorf("response row %d has non-numeric contribution count %T", i, values[i])
		}
		contribs = append(contribs, Contribution{name, "", int(value), -1})
	}
	return contribs, nil
}
