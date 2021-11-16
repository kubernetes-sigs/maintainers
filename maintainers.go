package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

type OwnersInfo struct {
	Filters   map[string]FiltersInfo `json:"filters,omitempty"`
	Approvers []string               `json:"approvers,omitempty"`
	Reviewers []string               `json:"reviewers,omitempty"`
}

type FiltersInfo struct {
	Approvers []string `json:"approvers,omitempty"`
	Reviewers []string `json:"reviewers,omitempty"`
}

// Aliases defines groups of people to be used in OWNERS files
type Aliases struct {
	RepoAliases map[string][]string `json:"aliases,omitempty"`
}

func main() {
	var fixupFlag bool
	pflag.BoolVarP(&fixupFlag, "fixup", "f", false, "Cleanup stale owner files")
	pflag.Parse()

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	userIDs := sets.String{}
	aliasPath, _ := filepath.Abs(filepath.Join(pwd, "OWNERS_ALIASES"))
	fmt.Printf("Processing %s\n", aliasPath)
	configAliases, err := getOwnerAliases(aliasPath)
	if err != nil {
		panic(err)
	}
	for _, ids := range configAliases.RepoAliases {
		InsertID(userIDs, ids...)
	}

	files, err := getOwnerFiles(pwd)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Printf("Processing %s\n", file)
		configOwners, err := getOwnersInfo(file)
		if err != nil {
			panic(err)
		}
		for _, filterInfo := range configOwners.Filters {
			InsertID(userIDs, filterInfo.Approvers...)
			InsertID(userIDs, filterInfo.Reviewers...)
		}
		InsertID(userIDs, configOwners.Approvers...)
		InsertID(userIDs, configOwners.Reviewers...)
	}

	for key, _ := range configAliases.RepoAliases {
		userIDs.Delete(key)
	}

	uniqueUsers := userIDs.List()
	fmt.Printf("Found %d unique aliases\n", len(configAliases.RepoAliases))
	fmt.Printf("Found %d unique users\n", len(uniqueUsers))

	err, contribs := getContributionsForAYear()
	if err != nil {
		panic(err)
	}
	var ownerContribs []Contribution
	for _, id := range uniqueUsers {
		for _, item := range contribs {
			if strings.ToLower(item.ID) == strings.ToLower(id) {
				ownerContribs = append(ownerContribs, Contribution{id, item.ID, item.Count})
				userIDs.Delete(id)
				break
			}
		}
	}

	fmt.Printf("\n\n>>>>> Contributions: %d\n", len(ownerContribs))
	sort.Slice(ownerContribs, func(i, j int) bool {
		return ownerContribs[i].Count > ownerContribs[j].Count
	})
	for _, item := range ownerContribs {
		if item.ID != item.alias {
			fmt.Printf("%s(%s) : %d\n", item.ID, item.alias, item.Count)
		} else {
			fmt.Printf("%s : %d\n", item.ID, item.Count)
		}
	}

	missingIDs := userIDs.List()
	sort.Strings(missingIDs)
	fmt.Printf("\n\n>>>>> Missing Contributions: %d\n", len(missingIDs))
	for _, id := range missingIDs {
		fmt.Printf("%#v\n", id)
	}

	if fixupFlag {
		files, err = getOwnerFiles(pwd)
		if err != nil {
			panic(err)
		}
		files = append(files, aliasPath)

		var regexArray []*regexp.Regexp
		for _, id := range missingIDs {
			searchRegex := regexp.MustCompile("- " + id)
			regexArray = append(regexArray, searchRegex)
		}
		for _, path := range files {
			fmt.Printf("Fixing up %s\n", path)
			data, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			contents := string(data)
			i := 0
			for _, regex := range regexArray {
				contents = regex.ReplaceAllString(contents, "# - " + missingIDs[i])
				i++
			}
			err = ioutil.WriteFile(path, []byte(contents), 0666)
			if err != nil {
				panic(err)
			}
		}
	}
}

func InsertID(s sets.String, ids ...string) {
	sort.Strings(ids)
	for _, id := range ids {
		s.Insert(id)
	}
}

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

type Frames struct {
	Schema map[string]interface{} `json:"schema,omitempty"`
	Data   Values                 `json:"data,omitempty"`
}

type Values struct {
	Items [][]interface{} `json:"values,omitempty"`
}

type Contribution struct {
	ID    string
	alias string
	Count int
}

func getContributionsForAYear() (error, []Contribution) {
	postBody := `{
	"queries": [{
		"refId": "A",
		"datasourceId": 1,
		"rawSql": "select sub.name as name, sub.value from (select row_number() over (order by sum(value) desc) as \"Rank\", split_part(name, '$$$', 1) as name, sum(value) as value from shdev where series = 'hdev_contributionsallall' and period = 'y' group by split_part(name, '$$$', 1)) sub",
		"format": "table"
	}]
}`
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
