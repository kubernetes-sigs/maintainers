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
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dims/maintainers/pkg/utils"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "ensure OWNERS, OWNERS_ALIASES and sigs.yaml have the correct data structure",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		aliasPath, err := utils.GetOwnersAliasesFile(pwd)
		if err == nil && len(aliasPath) > 0 {
			_, err := utils.GetOwnerAliases(aliasPath)
			if err != nil {
				panic(fmt.Errorf("error parsing file: %s - %w", aliasPath, err))
			}
		}

		var context *utils.Context
		sigsYamlPath, err := utils.GetSigsYamlFile(pwd)
		if err == nil && len(sigsYamlPath) > 0 {
			context, err = utils.GetSigsYaml(sigsYamlPath)
			if err != nil {
				panic(fmt.Errorf("error parsing file: %s - %w", sigsYamlPath, err))
			}
		}

		files, err := utils.GetOwnerFiles(pwd)
		if err != nil {
			panic(err)
		}

		for _, path := range files {
			_, err := utils.GetOwnersInfo(path)
			if err != nil {
				panic(fmt.Errorf("error parsing file: %s - %w", path, err))
			}
		}

		groupMap := context.PrefixToGroupMap()
		fileMap, errors := validateOwnersFilesInGroups(&groupMap)
		errors2 := warnFileMismatchesBetweenKubernetesRepoAndSigsYaml(fileMap)
		errors = append(errors, errors2...)

		if errors != nil && len(errors) > 0 {
			for _, err := range errors {
				fmt.Printf("WARNING: %v\n", err)
			}
			//panic("please see errors above")
		}
	},
}

func warnFileMismatchesBetweenKubernetesRepoAndSigsYaml(fileMap map[string]string) []error {
	var errors []error
	ownerFiles, err := utils.GetKubernetesOwnersFiles()
	if err != nil {
		panic(err)
	}

	for key, val := range fileMap {
		if strings.Index(key, "kubernetes/kubernetes") == -1 {
			continue
		}
		found := false
		for _, file := range *ownerFiles {
			if len(file) == 0 {
				continue
			}
			if strings.HasSuffix(key, file) {
				found = true
			}
		}
		if !found {
			errors = append(errors, fmt.Errorf("file [%s] in section %v is not present in kubernetes/kubernetes", key, val))
		}
	}

	for _, file := range *ownerFiles {
		if len(file) > 0 {
			if _, ok := fileMap[file]; !ok {
				errors = append(errors, fmt.Errorf("file [%s] is not in sigs.yaml", file))
			}
		}
	}

	return errors
}

func validateOwnersFilesInGroups(groupMap *map[string][]utils.Group) (map[string]string, []error) {
	fileMap := map[string]string{}
	var errors []error
	for groupType, groups := range *groupMap {
		for _, group := range groups {
			for _, sub := range group.Subprojects {
				for _, filePath := range sub.Owners {
					where := fmt.Sprintf("'%s/%s/%s'", groupType, group.Dir, sub.Name)
					if val, ok := fileMap[filePath]; ok {
						errors = append(errors, fmt.Errorf("%s is duplicated in %s and %s", filePath, val, where))
					} else {
						fileMap[filePath] = where
					}
				}
			}
		}
	}
	return fileMap, errors
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
