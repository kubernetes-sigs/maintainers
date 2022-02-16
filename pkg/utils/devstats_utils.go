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

	"github.com/pkg/errors"
)

func GetContributionsForAYear(repository string, period string) ([]Contribution, error) {
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
		return nil, errors.Wrap(err, fmt.Sprintf("bad error code from devstats: %d", resp.StatusCode))
	}

	var parsed map[string]map[string]map[string][]Frames
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse json from devstats")
	}

	foo := parsed["results"]["A"]["frames"][0].Data.Items[0]
	bar := parsed["results"]["A"]["frames"][0].Data.Items[1]

	var contribs []Contribution
	for i := 0; i < len(foo); i++ {
		contribs = append(contribs, Contribution{foo[i].(string), "", int(bar[i].(float64)), -1})
	}
	return contribs, nil
}
