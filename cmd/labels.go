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
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/dims/maintainers/pkg/utils"
)

// labelsCmd represents the export command
var labelsCmd = &cobra.Command{
	Use:   "labels",
	Short: "print a list of OWNERS files for labels",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		err = printFilesForLabels(pwd)
		if err != nil {
			panic(err)
		}
	},
}

var labelsFile string

func init() {
	labelsCmd.Flags().StringVar(&labelsFile, "output", "labels.csv", "output file path")
	labelsCmd.SilenceErrors = true
	rootCmd.AddCommand(labelsCmd)
}

func printFilesForLabels(pwd string) error {
	labelFiles := map[string]sets.String{}

	files, err := utils.GetOwnerFiles(pwd)
	if err != nil {
		return err
	}

	for _, file := range files {
		configOwners, err := utils.GetOwnersInfo(file)
		if err != nil {
			return fmt.Errorf("error processing %s: %w", file, err)
		}
		for _, label := range configOwners.Labels {
			var ok bool
			var val sets.String
			if val, ok = labelFiles[label]; !ok {
				val = sets.String{}
				labelFiles[label] = val
			}
			val.Insert(file)
		}
	}

	var labels []string
	for label := range labelFiles {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	for _, label := range labels {
		fmt.Printf("%s:\n", label)
		for _, file := range labelFiles[label].List() {
			fmt.Printf("\t%s\n", file)
		}
	}

	fmt.Printf("\n\n>>>>> generating %s\n", labelsFile)
	f, err := os.Create(labelsFile)
	if err != nil {
		return err
	}
	for _, label := range labels {
		for _, file := range labelFiles[label].List() {
			_, err := fmt.Fprintf(f, "%s,%s\n", label, file)
			if err != nil {
				return err
			}
		}
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}

	return nil
}
