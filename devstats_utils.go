package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func getContributionsForAYear(repository string) (error, []Contribution) {
	postBody := `{
	"queries": [{
		"refId": "A",
		"datasourceId": 1,
		"rawSql": "select sub.name as name, sub.value from (select row_number() over (order by sum(value) desc) as \"Rank\", split_part(name, '$$$', 1) as name, sum(value) as value from shdev_repos where series = 'hdev_contributionsallall' and period = 'y' group by split_part(name, '$$$', 1)) sub",
		"format": "table"
	}]
}`
	postBody = strings.Replace(
		postBody,
		"hdev_contributionsallall",
		fmt.Sprintf("hdev_contributions%sall", strings.ReplaceAll(repository, "/", "")),
		-1)

	requestBody := bytes.NewBuffer([]byte(postBody))
	resp, err := http.Post("https://k8s.devstats.cncf.io/api/ds/query", "application/json", requestBody)
	if err != nil {
		return err, nil
	}

	defer resp.Body.Close()
	var parsed map[string]map[string]map[string][]Frames
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return err, nil
	}

	foo := parsed["results"]["A"]["frames"][0].Data.Items[0]
	bar := parsed["results"]["A"]["frames"][0].Data.Items[1]

	var contribs []Contribution
	for i := 0; i < len(foo); i++ {
		contribs = append(contribs, Contribution{foo[i].(string), "", int(bar[i].(float64))})
	}
	return nil, contribs
}

