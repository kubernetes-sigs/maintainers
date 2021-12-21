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

var dryRun, skipGH, skipDS, export bool
var repositoryDS, repositoryGH string
var includes, excludes []string

func init() {
	pflag.StringSliceVar(&includes, "include", []string{}, "add these comma-separated list of users to prune from OWNERS")
	pflag.StringSliceVar(&excludes, "exclude", []string{}, "do not prune these comma-separated list of users from OWNERS")
	pflag.BoolVar(&dryRun, "dryrun", true, "do not modify any files")
	pflag.BoolVar(&skipGH, "skip-github", false, "skip github PR count check")
	pflag.BoolVar(&skipDS, "skip-devstats", false, "skip devstat contributions count check")
	pflag.StringVar(&repositoryDS, "repository-devstats", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
	pflag.StringVar(&repositoryGH, "repository-github", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
	pflag.BoolVar(&export, "export", false, "export contents of all owners related files as output.csv")
}

func main() {
	pflag.Parse()

	fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	userIDs, repoAliases, files, err := getOwnersAndAliases(pwd)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Printf("Processed %s\n", file)
	}
	uniqueUsers := userIDs.List()
	fmt.Printf("Found %d unique aliases\n", len(repoAliases))
	fmt.Printf("Found %d unique users\n", len(uniqueUsers))

	var ownerContribs []Contribution

	if !skipDS {
		err, contribs := getContributionsForAYear(repositoryDS)
		if err != nil {
			panic(err)
		}
		if contribs == nil || len(contribs) == 0 {
			panic("unable to find any contributions in repository : " + repositoryDS)
		}
		for _, id := range uniqueUsers {
			for _, item := range contribs {
				if strings.ToLower(item.ID) == strings.ToLower(id) {
					ownerContribs = append(ownerContribs, Contribution{id, item.ID, item.ContribCount, -1})
					userIDs.Delete(id)
					break
				}
			}
		}
	} else {
		for _, id := range uniqueUsers {
			ownerContribs = append(ownerContribs, Contribution{id, id, -1, -1})
			userIDs.Delete(id)
		}
	}

	var lowPRComments []string
	if !skipGH {
		lowPRComments = fetchGithubPRCommentCounts(ownerContribs)
	}

	// Sort by descending order of contributions/comments in devstats
	sort.Slice(ownerContribs, func(i, j int) bool {
		return ownerContribs[i].ContribCount > ownerContribs[j].ContribCount &&
			ownerContribs[i].CommentCount > ownerContribs[j].CommentCount
	})

	fmt.Printf("\n\n>>>>> Contributions from %s devstats repo and %s github repo : %d\n", repositoryDS, repositoryGH, len(ownerContribs))
	fmt.Printf(">>>>> GitHub ID : Devstats contrib count : GitHub PR comment count\n")
	for _, item := range ownerContribs {
		if item.ID != item.alias {
			fmt.Printf("%s(%s) : %d : %d \n", item.ID, item.alias, item.ContribCount, item.CommentCount)
		} else {
			fmt.Printf("%s : %d : %d \n", item.ID, item.ContribCount, item.CommentCount)
		}
	}

	missingIDs := userIDs.List()
	sort.Strings(missingIDs)
	if !skipDS {
		fmt.Printf("\n\n>>>>> Missing Contributions in %s (devstats == 0): %d\n", repositoryDS, len(missingIDs))
		for _, id := range missingIDs {
			fmt.Printf("%s\n", id)
		}
	}

	if !skipGH {
		fmt.Printf("\n\n>>>>> Low reviews/approvals in %s (GH pr comments <= 10 && devstats <=20): %d\n",
			repositoryGH, len(lowPRComments))
		for _, id := range lowPRComments {
			fmt.Printf("%s\n", id)
		}
	}

	if !dryRun {
		err = fixupOwnersFiles(files, missingIDs, lowPRComments)
		if err != nil {
			panic(err)
		}
	}

	if export {
		err = exportOwnersAndAliases(pwd)
		if err != nil {
			panic(err)
		}
	}
}

func fetchGithubPRCommentCounts(ownerContribs []Contribution) []string {
	var lowPRComments []string
	var commentCount int
	for count, item := range ownerContribs {
		commentCount = -1
		var err error
		commentCount, err = fetchPRCommentCount(item.ID, repositoryGH)
		for commentCount == -1 && err == nil {
			fmt.Printf(".")
			time.Sleep(5 * time.Second)
			commentCount, err = fetchPRCommentCount(item.ID, repositoryGH)
		}
		if item.ContribCount <= 20 && commentCount <= 10 {
			lowPRComments = append(lowPRComments, item.ID)
		}
		fmt.Printf(".")
		time.Sleep(2 * time.Second)
		ownerContribs[count].CommentCount = commentCount
	}
	fmt.Printf("\n")
	return lowPRComments
}

func fixupOwnersFiles(files []string, missingIDs []string, lowPRComments []string) error {
	userIDs := sets.String{}

	userIDs.Insert(missingIDs...)
	userIDs.Insert(lowPRComments...)
	userIDs.Insert(includes...)
	userIDs.Delete(excludes...)

	list := userIDs.List()
	for _, path := range files {
		err := removeUserFromOWNERS(path, list)
		if err != nil {
			return err
		}
	}
	return nil
}

func getOwnersAndAliases(pwd string) (sets.String, map[string][]string, []string, error) {
	userIDs := sets.String{}
	var repoAliases map[string][]string
	aliasPath, _ := filepath.Abs(filepath.Join(pwd, "OWNERS_ALIASES"))
	if _, err := os.Stat(aliasPath); err == nil {
		configAliases, err := getOwnerAliases(aliasPath)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, ids := range configAliases.RepoAliases {
			userIDs.Insert(ids...)
		}
		repoAliases = configAliases.RepoAliases
	}

	files, err := getOwnerFiles(pwd)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, file := range files {
		configOwners, err := getOwnersInfo(file)
		if err != nil {
			return nil, nil, nil, err
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
	return userIDs, repoAliases, files, nil
}

func exportOwnersAndAliases(pwd string) error {
	var repoAliases map[string][]string

	aliasPath, _ := filepath.Abs(filepath.Join(pwd, "OWNERS_ALIASES"))
	if _, err := os.Stat(aliasPath); err == nil {
		configAliases, err := getOwnerAliases(aliasPath)
		if err != nil {
			return err
		}
		repoAliases = configAliases.RepoAliases
	}

	files, err := getOwnerFiles(pwd)
	if err != nil {
		return err
	}

	type Row struct {
		id    string
		alias string
		file  string
	}
	var rows []Row

	for _, file := range files {
		userIDs := sets.String{}
		aliases := sets.String{}
		configOwners, err := getOwnersInfo(file)
		if err != nil {
			return err
		}
		for _, filterInfo := range configOwners.Filters {
			userIDs.Insert(filterInfo.Approvers...)
			userIDs.Insert(filterInfo.Reviewers...)
		}
		userIDs.Insert(configOwners.Approvers...)
		userIDs.Insert(configOwners.Reviewers...)
		for key, _ := range repoAliases {
			if userIDs.Has(key) {
				userIDs.Delete(key)
				aliases.Insert(key)
			}
		}

		for _, id := range userIDs.List() {
			rows = append(rows, Row{id, "", file})
		}
		for _, alias := range aliases.List() {
			ids, ok := repoAliases[alias]
			if ok {
				for _, id := range ids {
					rows = append(rows, Row{id, alias, file})
				}
			}
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		switch strings.Compare(rows[i].id, rows[j].id) {
		case -1:
			return true
		case 1:
			return false
		}
		switch strings.Compare(rows[i].alias, rows[j].alias) {
		case -1:
			return true
		case 1:
			return false
		}
		return rows[i].file > rows[j].file
	})
	fmt.Printf("\n\n>>>>> generating export.csv\n")
	f, err := os.Create("export.csv")
	if err != nil {
		panic(err)
	}
	for _, row := range rows {
		_, err = fmt.Fprintf(f, "%s,%s,%s\n", row.id, row.alias, row.file)
		if err != nil {
			return err
		}
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}

	return nil
}
