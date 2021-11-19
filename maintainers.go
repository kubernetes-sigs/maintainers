package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"

	yaml3 "gopkg.in/yaml.v3"
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
	fmt.Printf(">>>>> GitHub ID : Devstats contrib count : GitHub PR comment count\n")
	sort.Slice(ownerContribs, func(i, j int) bool {
		return ownerContribs[i].Count > ownerContribs[j].Count
	})
	var lowPRComments []string
	for _, item := range ownerContribs {
		commentCount, err := fetchPRCommentCount(item.ID)
		for commentCount == -1 && err == nil {
			time.Sleep(5 * time.Second)
			commentCount, err = fetchPRCommentCount(item.ID)
		}
		if item.ID != item.alias {
			fmt.Printf("%s(%s) : %d : %d \n", item.ID, item.alias, item.Count, commentCount)
		} else {
			fmt.Printf("%s : %d : %d \n", item.ID, item.Count, commentCount)
		}
		if item.Count <= 20 && commentCount <= 10 {
			lowPRComments = append(lowPRComments, item.ID)
		}
		time.Sleep(2 * time.Second)
	}

	missingIDs := userIDs.List()
	sort.Strings(missingIDs)
	fmt.Printf("\n\n>>>>> Missing Contributions (devstats == 0): %d\n", len(missingIDs))
	for _, id := range missingIDs {
		fmt.Printf("%#v\n", id)
	}

	fmt.Printf("\n\n>>>>> Low reviews/approvals (GH pr comments <= 10 && devstats <=20): %d\n", len(missingIDs))
	for _, id := range lowPRComments {
		fmt.Printf("%#v\n", id)
	}

	if fixupFlag {
		files, err = getOwnerFiles(pwd)
		if err != nil {
			panic(err)
		}
		files = append(files, aliasPath)
		for _, path := range files {
			err = removeUserFromOWNERS(path, missingIDs)
			if err != nil {
				panic(err)
			}
		}
	}
}

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

func removeUserFromOWNERS(path string, users []string) error {
	fmt.Printf("Fixing up %s\n", path)
	for _, user := range users {
		sourceYaml, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		rootNode := yaml3.Node{}
		err = yaml3.Unmarshal(sourceYaml, &rootNode)
		if err != nil {
			return err
		}

		switchToEmeritus(&rootNode, user)

		writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		encoder := yaml3.NewEncoder(writer)
		encoder.SetIndent(2)
		err = encoder.Encode(&rootNode)
		if err != nil {
			return err
		}
		err = encoder.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func switchToEmeritus(rootNode *yaml3.Node, user string) {
	// find mapping node
	mappingNode := fetchMappingNode(rootNode)
	if mappingNode == nil {
		return
	}

	// cleanup user from approvers and reviewers
	if !removeUserFromApproversAndReviewers(mappingNode, user) {
		return
	}

	// add user to emeritus list, create things we need if they are not there already
	emeritusSeqNode := fetchEmeritusSequenceNode(mappingNode)

	// add if not already present
	addUserToEmeritusList(emeritusSeqNode, user)
}

func addUserToEmeritusList(emeritusSeqNode *yaml3.Node, user string) {
	foundInEmeritusList := false
	for _, item := range emeritusSeqNode.Content {
		if item.Kind == yaml3.ScalarNode && strings.ToLower(item.Value) == strings.ToLower(user) {
			foundInEmeritusList = true
		}
	}
	if !foundInEmeritusList {
		node := yaml3.Node{
			Kind:  yaml3.ScalarNode,
			Tag:   "!!str",
			Value: user,
		}
		emeritusSeqNode.Content = append(emeritusSeqNode.Content, &node)
	}
}

func fetchEmeritusSequenceNode(mappingNode *yaml3.Node) *yaml3.Node {
	var emeritusScalar, emeritusSeqNode *yaml3.Node
	for i := 0; i < len(mappingNode.Content); i++ {
		node := mappingNode.Content[i]
		if node.Kind == yaml3.ScalarNode && (node.Value == "emeritus_approvers") {
			emeritusScalar = node
			j := i + 1
			if j < len(mappingNode.Content) && mappingNode.Content[j].Kind == yaml3.SequenceNode {
				emeritusSeqNode = mappingNode.Content[j]
			}
		}
	}
	if emeritusScalar == nil {
		emeritusScalar = &yaml3.Node{
			Kind:  yaml3.ScalarNode,
			Tag:   "!!str",
			Value: "emeritus_approvers",
		}
		mappingNode.Content = append(mappingNode.Content, emeritusScalar)
	}
	if emeritusSeqNode == nil {
		// remove trailing null scalar
		lastNode := mappingNode.Content[len(mappingNode.Content)-1]
		if lastNode.Kind == yaml3.ScalarNode && lastNode.Tag == "!!null" {
			mappingNode.Content = mappingNode.Content[:len(mappingNode.Content)-1]
		}
		emeritusSeqNode = &yaml3.Node{
			Kind: yaml3.SequenceNode,
			Tag:  "!!seq",
		}
		mappingNode.Content = append(mappingNode.Content, emeritusSeqNode)
	}
	return emeritusSeqNode
}

func fetchMappingNode(rootNode *yaml3.Node) *yaml3.Node {
	var mappingNode *yaml3.Node
	for _, node := range rootNode.Content {
		if node.Kind == yaml3.MappingNode {
			mappingNode = node
			break
		}
	}
	if mappingNode != nil {
		// structure of the file is slightly different for OWNERS_ALIASES
		foundAliases := false
		for _, node := range mappingNode.Content {
			if node.Kind == yaml3.ScalarNode && node.Value == "aliases" {
				foundAliases = true
			}
		}
		if foundAliases {
			for _, node := range mappingNode.Content {
				if node.Kind == yaml3.MappingNode {
					return node
				}
			}
		}
	}
	return mappingNode
}

func removeUserFromApproversAndReviewers(mappingNode *yaml3.Node, user string) bool {
	foundInApproverList := false
	for i := 0; i < len(mappingNode.Content); i++ {
		node := mappingNode.Content[i]
		if node.Kind == yaml3.ScalarNode && node.Value != "emeritus_approvers" {
			// found the approvers or reviewer node
			j := i + 1
			if j < len(mappingNode.Content) && mappingNode.Content[j].Kind == yaml3.SequenceNode {
				seqNode := mappingNode.Content[j]
				// found the list of approvers or reviewers
				var newList []*yaml3.Node
				for _, item := range seqNode.Content {
					// skip the user we want to eliminate from the list
					if item.Kind == yaml3.ScalarNode && strings.ToLower(item.Value) == strings.ToLower(user) {
						if node.Value == "approvers" {
							foundInApproverList = true
						}
					} else {
						newList = append(newList, item)
					}
				}
				seqNode.Content = newList
			}
		}
	}
	return foundInApproverList
}
