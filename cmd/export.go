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
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/dims/maintainers/pkg/utils"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export contents of OWNERS and OWNERS_ALIASES as parsable csv file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		err = exportOwnersAndAliases(pwd)
		if err != nil {
			panic(err)
		}
	},
}

var outputFile string

func init() {
	exportCmd.Flags().StringVar(&outputFile, "output", "export.csv", "output file path")
	exportCmd.SilenceErrors = true
	rootCmd.AddCommand(exportCmd)
}

func exportOwnersAndAliases(pwd string) error {
	var repoAliases map[string][]string

	aliasPath, err := utils.GetOwnersAliasesFile(pwd)
	if err == nil && len(aliasPath) > 0 {
		configAliases, err := utils.GetOwnerAliases(aliasPath)
		if err != nil {
			return err
		}
		repoAliases = configAliases.RepoAliases
	}

	files, err := utils.GetOwnerFiles(pwd)
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
		configOwners, err := utils.GetOwnersInfo(file)
		if err != nil {
			return fmt.Errorf("error processing %s: %w", file, err)
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
	fmt.Printf("\n\n>>>>> generating %s\n", outputFile)
	f, err := os.Create(outputFile)
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
