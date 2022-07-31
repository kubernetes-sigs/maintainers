/*
Copyright 2022 The Kubernetes Authors.

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

	"github.com/kubernetes-sigs/maintainers/pkg/utils"
	"github.com/spf13/cobra"
)

// emeritusStatsCmd represents the emeritus-stats command
var emeritusStatsCmd = &cobra.Command{
	Use:   "emeritus-stats",
	Short: "emeritus-stats gives stats on churn around emeritus_approvers in a specified time frame",
	Long: `emeritus-stats outputs how many emeritus_approvers or
emeritus_reviewers were added or removed in a specified time frame
across all OWNERS files of a specified repository.

Along with this, it also outputs the average number of emeritus_approvers
or emeritus_reviewers added or removed per OWNERS file.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return flags.validate()
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		// Get the current branch in order to restore it later.
		currBranch, err := utils.GetBranchName(flags.dir)
		if err != nil {
			return err
		}

		defer func() {
			err = utils.Checkout(currBranch, flags.dir)
		}()

		// Checkout the repo at the from date.
		err = utils.CheckoutAtDate(flags.branch, flags.from, flags.dir)
		if err != nil {
			return err
		}

		// Get EmeritusCounts for the from date.
		fromCounts, err := utils.GetEmeritusCounts(flags.dir)
		if err != nil {
			return err
		}

		// Checkout the repo at the to date.
		err = utils.CheckoutAtDate(flags.branch, flags.to, flags.dir)
		if err != nil {
			return err
		}

		// Get EmeritusCounts for the to date.
		toCounts, err := utils.GetEmeritusCounts(flags.dir)
		if err != nil {
			return err
		}

		// Calculate the difference.
		diff := utils.CalculateEmeritusDiff(fromCounts, toCounts)

		fmt.Printf("Info for the time period: %s - %s\n\n", flags.from, flags.to)
		fmt.Printf("For emeritus_approvers:\n\n")
		diff.Approvers.PrettyPrint()
		fmt.Printf("\nFor emeritus_reviewers:\n\n")
		diff.Reviewers.PrettyPrint()

		return nil
	},
}

type cmdFlags struct {
	from, to, dir, branch string
}

var flags = cmdFlags{}

func (f cmdFlags) validate() error {
	if len(f.from) == 0 {
		return fmt.Errorf("from date needs to be specified")
	}
	if len(f.to) == 0 {
		return fmt.Errorf("to date needs to be specified")
	}
	if len(f.dir) == 0 {
		return fmt.Errorf("dir needs to be specified")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(emeritusStatsCmd)
	emeritusStatsCmd.SilenceErrors = true

	emeritusStatsCmd.Flags().StringVarP(&flags.from, "from", "f", "", "from date in format yyyy-mm-dd")
	emeritusStatsCmd.Flags().StringVarP(&flags.to, "to", "t", "", "to date in format yyyy-mm-dd")
	emeritusStatsCmd.Flags().StringVarP(&flags.dir, "dir", "d", "", "local directory where the repo is")
	// Defaulting to master considering this is going to be run on k/k more than other repositories.
	emeritusStatsCmd.
		Flags().
		StringVarP(&flags.branch, "branch", "b", "master", "base branch on which checkout should be done")
}
