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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubernetes-sigs/maintainers/pkg/utils"
)

type options struct {
	dryRun       bool
	skipDS       bool
	skipGH       bool
	repositoryDS string
	repositoryGH string
	periodDS     string
	includes     []string
	excludes     []string
	excludeFiles []string
}

var o options

func init() {
	pruneCmd.Flags().StringSliceVar(&o.includes, "include", []string{}, "add these comma-separated list of users to prune from OWNERS")
	pruneCmd.Flags().StringSliceVar(&o.excludes, "exclude", []string{}, "do not prune these comma-separated list of users from OWNERS")
	pruneCmd.Flags().BoolVar(&o.dryRun, "dryrun", true, "do not modify any files")
	pruneCmd.Flags().BoolVar(&o.skipGH, "skip-github", false, "skip github PR count check")
	pruneCmd.Flags().BoolVar(&o.skipDS, "skip-devstats", false, "skip devstat contributions count check")
	pruneCmd.Flags().StringVar(&o.repositoryDS, "repository-devstats", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
	pruneCmd.Flags().StringVar(&o.repositoryGH, "repository-github", "kubernetes/kubernetes", "defaults to \"kubernetes/kubernetes\" repository")
	pruneCmd.Flags().StringVar(&o.periodDS, "period-devstats", "y", "one of \"y\" (year) \"q\" (quarter) \"m\" (month) ")
	pruneCmd.Flags().StringSliceVar(&o.excludeFiles, "exclude-files", []string{}, "do not update these OWNERS files")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	pruneCmd.SilenceErrors = true
	rootCmd.AddCommand(pruneCmd)
}

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove stale github ids from OWNERS and OWNERS_ALIASES",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		userIDs, repoAliases, files, err := getOwnersAndAliases(pwd)
		if err != nil {
			return err
		}
		for _, file := range files {
			fmt.Printf("Processed %s\n", file)
		}
		uniqueUsers := userIDs.List()
		fmt.Printf("Found %d unique aliases\n", len(repoAliases))
		fmt.Printf("Found %d unique users\n", len(uniqueUsers))

		var ownerContribs []utils.Contribution

		if !o.skipDS {
			contribs, err := utils.GetContributionsForAYear(o.repositoryDS, o.periodDS)
			if err != nil {
				return err
			}
			if len(contribs) == 0 {
				panic("unable to find any contributions in repository : " + o.repositoryDS)
			}
			for _, id := range uniqueUsers {
				for _, item := range contribs {
					if strings.EqualFold(item.ID, id) {
						ownerContribs = append(ownerContribs,
							utils.Contribution{
								ID:           id,
								Alias:        item.ID,
								ContribCount: item.ContribCount,
								CommentCount: -1,
							},
						)
						userIDs.Delete(id)
						break
					}
				}
			}
		} else {
			for _, id := range uniqueUsers {
				ownerContribs = append(ownerContribs,
					utils.Contribution{
						ID:           id,
						Alias:        id,
						ContribCount: -1,
						CommentCount: -1,
					},
				)
				userIDs.Delete(id)
			}
		}

		var lowPRComments []string
		if !o.skipGH {
			lowPRComments = fetchGithubPRCommentCounts(ownerContribs)
		}

		// Sort by descending order of contributions/comments in devstats
		sort.Slice(ownerContribs, func(i, j int) bool {
			return ownerContribs[i].ContribCount > ownerContribs[j].ContribCount &&
				ownerContribs[i].CommentCount > ownerContribs[j].CommentCount
		})

		fmt.Printf("\n\n>>>>> Contributions from %s devstats repo and %s github repo : %d\n", o.repositoryDS, o.repositoryGH, len(ownerContribs))
		fmt.Printf(">>>>> GitHub ID : Devstats contrib count : GitHub PR comment count\n")
		for _, item := range ownerContribs {
			if item.ID != item.Alias {
				fmt.Printf("%s(%s) : %d : %d \n", item.ID, item.Alias, item.ContribCount, item.CommentCount)
			} else {
				fmt.Printf("%s : %d : %d \n", item.ID, item.ContribCount, item.CommentCount)
			}
		}

		missingIDs := userIDs.List()
		sort.Strings(missingIDs)
		if !o.skipDS {
			fmt.Printf("\n\n>>>>> Missing Contributions in %s (devstats == 0): %d\n", o.repositoryDS, len(missingIDs))
			for _, id := range missingIDs {
				fmt.Printf("%s\n", id)
			}
		}

		if !o.skipGH {
			fmt.Printf("\n\n>>>>> Low reviews/approvals in %s (GH pr comments <= 10 && devstats <=20): %d\n",
				o.repositoryGH, len(lowPRComments))
			for _, id := range lowPRComments {
				fmt.Printf("%s\n", id)
			}
		}

		if !o.dryRun {
			err = fixupOwnersFiles(files, missingIDs, lowPRComments)
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("--dryrun is set to true, will skip updating OWNERS and OWNER_ALIASES")
		}
		return nil
	},
}

func fetchGithubPRCommentCounts(ownerContribs []utils.Contribution) []string {
	var lowPRComments []string
	var commentCount int
	for count, item := range ownerContribs {
		commentCount = -1
		var err error
		commentCount, err = utils.FetchPRCommentCount(item.ID, o.repositoryGH)
		for commentCount == -1 && err == nil {
			fmt.Printf(".")
			time.Sleep(5 * time.Second)
			commentCount, err = utils.FetchPRCommentCount(item.ID, o.repositoryGH)
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
	userIDs.Insert(o.includes...)
	userIDs.Delete(o.excludes...)

	list := userIDs.List()
	for _, path := range files {
		if isExcludedPath(path, o.excludeFiles) {
			continue
		}
		err := utils.RemoveUserFromOWNERS(path, list)
		if err != nil {
			return err
		}
	}
	return nil
}

func isExcludedPath(a string, list []string) bool {
	for _, b := range list {
		pathB, _ := filepath.Abs(b)
		if pathB == a {
			return true
		}
	}
	return false
}

func getOwnersAndAliases(pwd string) (sets.String, map[string][]string, []string, error) {
	userIDs := sets.String{}
	var repoAliases map[string][]string
	aliasPath, err := utils.GetOwnersAliasesFile(pwd)
	if err == nil && len(aliasPath) > 0 {
		configAliases, err := utils.GetOwnerAliases(aliasPath)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, ids := range configAliases.RepoAliases {
			userIDs.Insert(ids...)
		}
		repoAliases = configAliases.RepoAliases
	}

	files, err := utils.GetOwnerFiles(pwd)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, file := range files {
		configOwners, err := utils.GetOwnersInfo(file)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error processing %s: %w", file, err)
		}
		for _, filterInfo := range configOwners.Filters {
			userIDs.Insert(filterInfo.Approvers...)
			userIDs.Insert(filterInfo.Reviewers...)
		}
		userIDs.Insert(configOwners.Approvers...)
		userIDs.Insert(configOwners.Reviewers...)
	}

	for key := range repoAliases {
		userIDs.Delete(key)
	}
	if len(aliasPath) > 0 {
		files = append(files, aliasPath)
	}
	return userIDs, repoAliases, files, nil
}
