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

var dryRun, skipGH bool
var repositoryDS, repositoryGH string

func init() {
	pflag.BoolVarP(&dryRun, "dryrun", "r", true, "do not modify any files")
	pflag.BoolVarP(&skipGH, "skip-github", "s", false, "skip github PR count check")
	pflag.StringVarP(&repositoryDS, "repository-devstats", "d", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
	pflag.StringVarP(&repositoryGH, "repository-github", "g", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
}

func main() {
	pflag.Parse()

	fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	userIDs, repoAliases, files := getOwnersAndAliases(pwd)
	for _, file := range files {
		fmt.Printf("Processed %s\n", file)
	}
	uniqueUsers := userIDs.List()
	fmt.Printf("Found %d unique aliases\n", len(repoAliases))
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
				ownerContribs = append(ownerContribs, Contribution{id, item.ID, item.ContribCount, -1})
				userIDs.Delete(id)
				break
			}
		}
	}

	// Sort by descending order of contributions in devstats
	sort.Slice(ownerContribs, func(i, j int) bool {
		return ownerContribs[i].ContribCount > ownerContribs[j].ContribCount
	})

	fmt.Printf("\n\n>>>>> Contributions from %s devstats repo and %s github repo : %d\n", repositoryDS, repositoryGH, len(ownerContribs))
	fmt.Printf(">>>>> GitHub ID : Devstats contrib count : GitHub PR comment count\n")
	lowPRComments := fetchGithubPRCommentCounts(ownerContribs, err)
	for _, item := range ownerContribs {
		if item.ID != item.alias {
			fmt.Printf("%s(%s) : %d : %d \n", item.ID, item.alias, item.ContribCount, item.CommentCount)
		} else {
			fmt.Printf("%s : %d : %d \n", item.ID, item.ContribCount, item.CommentCount)
		}
	}

	missingIDs := userIDs.List()
	sort.Strings(missingIDs)
	fmt.Printf("\n\n>>>>> Missing Contributions in %s (devstats == 0): %d\n", repositoryDS, len(missingIDs))
	for _, id := range missingIDs {
		fmt.Printf("%s\n", id)
	}

	if !skipGH {
		fmt.Printf("\n\n>>>>> Low reviews/approvals in %s (GH pr comments <= 10 && devstats <=20): %d\n",
			repositoryGH, len(lowPRComments))
		for _, id := range lowPRComments {
			fmt.Printf("%s\n", id)
		}
	}

	if !dryRun {
		fixupOwnersFiles(files, missingIDs, lowPRComments)
	}
}

func fetchGithubPRCommentCounts(ownerContribs []Contribution, err error) []string {
	var lowPRComments []string
	var commentCount int
	for count, item := range ownerContribs {
		commentCount = -1
		if !skipGH {
			commentCount, err = fetchPRCommentCount(item.ID, repositoryGH)
			for commentCount == -1 && err == nil {
				time.Sleep(5 * time.Second)
				commentCount, err = fetchPRCommentCount(item.ID, repositoryGH)
			}
			if item.ContribCount <= 20 && commentCount <= 10 {
				lowPRComments = append(lowPRComments, item.ID)
			}
			time.Sleep(2 * time.Second)
		}
		ownerContribs[count].CommentCount = commentCount
	}
	return lowPRComments
}

func fixupOwnersFiles(files []string, missingIDs []string, lowPRComments []string) {
	for _, path := range files {
		err := removeUserFromOWNERS(path, missingIDs)
		if err != nil {
			panic(err)
		}
	}
	for _, path := range files {
		err := removeUserFromOWNERS(path, lowPRComments)
		if err != nil {
			panic(err)
		}
	}
}

func getOwnersAndAliases(pwd string) (sets.String, map[string][]string, []string) {
	userIDs := sets.String{}
	var repoAliases map[string][]string
	aliasPath, _ := filepath.Abs(filepath.Join(pwd, "OWNERS_ALIASES"))
	if _, err := os.Stat(aliasPath); err == nil {
		configAliases, err := getOwnerAliases(aliasPath)
		if err != nil {
			panic(err)
		}
		for _, ids := range configAliases.RepoAliases {
			userIDs.Insert(ids...)
		}
		repoAliases = configAliases.RepoAliases
	}

	files, err := getOwnerFiles(pwd)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
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
	files = append(files, aliasPath)
	return userIDs, repoAliases, files
}
