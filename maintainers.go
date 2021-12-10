package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/sets"
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
	var fixupFlag, skipGH bool
	var repository string
	pflag.BoolVarP(&fixupFlag, "fixup", "f", true, "Cleanup stale owner files")
	pflag.BoolVarP(&skipGH, "skipGH", "s", false, "skip github PR count check")
	pflag.StringVarP(&repository, "repository", "r", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
	pflag.Parse()

	fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
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

	err, contribs := getContributionsForAYear(repository)
	if err != nil {
		panic(err)
	}
	if contribs == nil || len(contribs) == 0 {
		panic("unable to find any contributions in repository : " + repository)
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
		commentCount := -1
		if !skipGH {
			commentCount, err = fetchPRCommentCount(item.ID, repository)
			for commentCount == -1 && err == nil {
				time.Sleep(5 * time.Second)
				commentCount, err = fetchPRCommentCount(item.ID, repository)
			}
			if item.Count <= 20 && commentCount <= 10 {
				lowPRComments = append(lowPRComments, item.ID)
			}
			time.Sleep(2 * time.Second)
		}
		if item.ID != item.alias {
			fmt.Printf("%s(%s) : %d : %d \n", item.ID, item.alias, item.Count, commentCount)
		} else {
			fmt.Printf("%s : %d : %d \n", item.ID, item.Count, commentCount)
		}
	}

	missingIDs := userIDs.List()
	sort.Strings(missingIDs)
	fmt.Printf("\n\n>>>>> Missing Contributions (devstats == 0): %d\n", len(missingIDs))
	for _, id := range missingIDs {
		fmt.Printf("%#v\n", id)
	}

	if !skipGH {
		fmt.Printf("\n\n>>>>> Low reviews/approvals (GH pr comments <= 10 && devstats <=20): %d\n", len(lowPRComments))
		for _, id := range lowPRComments {
			fmt.Printf("%#v\n", id)
		}
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

func InsertID(s sets.String, ids ...string) {
	sort.Strings(ids)
	for _, id := range ids {
		s.Insert(id)
	}
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
