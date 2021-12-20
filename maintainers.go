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

func main() {
	var dryRun, skipGH bool
	var repositoryDS, repositoryGH string
	pflag.BoolVarP(&dryRun, "dryrun", "r", true, "do not modify any files")
	pflag.BoolVarP(&skipGH, "skip-github", "s", false, "skip github PR count check")
	pflag.StringVarP(&repositoryDS, "repository-devstats", "d", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
	pflag.StringVarP(&repositoryGH, "repository-github", "g", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
	pflag.Parse()

	fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	userIDs := sets.String{}
	var repoAliases map[string][]string
	aliasPath, _ := filepath.Abs(filepath.Join(pwd, "OWNERS_ALIASES"))
	if _, err := os.Stat(aliasPath); err == nil {
		fmt.Printf("Processing %s\n", aliasPath)
		configAliases, err := getOwnerAliases(aliasPath)
		if err != nil {
			panic(err)
		}
		for _, ids := range configAliases.RepoAliases {
			userIDs.Insert(ids...)
		}
		repoAliases = configAliases.RepoAliases
		fmt.Printf("Found %d unique aliases\n", len(repoAliases))
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
			userIDs.Insert(filterInfo.Approvers...)
			userIDs.Insert(filterInfo.Reviewers...)
		}
		userIDs.Insert(configOwners.Approvers...)
		userIDs.Insert(configOwners.Reviewers...)
	}

	for key, _ := range repoAliases {
		userIDs.Delete(key)
	}

	uniqueUsers := userIDs.List()
	fmt.Printf("Found %d unique users\n", len(uniqueUsers))

	err, contribs := getContributionsForAYear(repositoryDS)
	if err != nil {
		panic(err)
	}
	if contribs == nil || len(contribs) == 0 {
		panic("unable to find any contributions in repository : " + repositoryDS)
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

	fmt.Printf("\n\n>>>>> Contributions from %s devstats repo and %s github repo : %d\n", repositoryDS, repositoryGH, len(ownerContribs))
	fmt.Printf(">>>>> GitHub ID : Devstats contrib count : GitHub PR comment count\n")
	sort.Slice(ownerContribs, func(i, j int) bool {
		return ownerContribs[i].Count > ownerContribs[j].Count
	})
	var lowPRComments []string
	for _, item := range ownerContribs {
		commentCount := -1
		if !skipGH {
			commentCount, err = fetchPRCommentCount(item.ID, repositoryGH)
			for commentCount == -1 && err == nil {
				time.Sleep(5 * time.Second)
				commentCount, err = fetchPRCommentCount(item.ID, repositoryGH)
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
		fmt.Printf("%s\n", id)
	}

	if !skipGH {
		fmt.Printf("\n\n>>>>> Low reviews/approvals (GH pr comments <= 10 && devstats <=20): %d\n", len(lowPRComments))
		for _, id := range lowPRComments {
			fmt.Printf("%s\n", id)
		}
	}

	if !dryRun {
		files, err = getOwnerFiles(pwd)
		if err != nil {
			panic(err)
		}
		if _, err := os.Stat(aliasPath); err == nil {
			files = append(files, aliasPath)
		}
		for _, path := range files {
			err = removeUserFromOWNERS(path, missingIDs)
			if err != nil {
				panic(err)
			}
		}
		for _, path := range files {
			err = removeUserFromOWNERS(path, lowPRComments)
			if err != nil {
				panic(err)
			}
		}
	}
}
