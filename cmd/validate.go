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
	"time"

	"github.com/spf13/cobra"

	"maintainers/pkg/utils"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "ensure OWNERS, OWNERS_ALIASES, sigs.yaml have the correct data structure",
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
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
