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
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var yamlFile string

func init() {
	checkURLsCmd.Flags().StringVar(&yamlFile, "yaml-file", "sigs.yaml", "validate urls in this yaml file")
	rootCmd.AddCommand(checkURLsCmd)
}

// exportCmd represents the export command
var checkURLsCmd = &cobra.Command{
	Use:   "check-urls",
	Short: "ensure all the urls in yaml file are still valid",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
		fmt.Printf("Processing %s\n", yamlFile)
		sourceYaml, err := ioutil.ReadFile(yamlFile)
		if err != nil {
			panic(err)
		}
		rootNode := yaml.Node{}
		err = yaml.Unmarshal(sourceYaml, &rootNode)
		if err != nil {
			panic(err)
		}
		ok := processNode(&rootNode)
		fmt.Println("done")
		if !ok {
			os.Exit(1)
		}
	},
}

func processNode(node *yaml.Node) (ret bool) {
	ret = true
	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" &&
		(strings.Index(node.Value, "https://") == 0 ||
			strings.Index(node.Value, "http://") == 0) {
		res, err := http.Head(node.Value)
		if err != nil || res.StatusCode != 200 {
			fmt.Printf("found invalid url: %s (http code: %d) at (%d,%d)",
				node.Value, res.StatusCode, node.Line, node.Column)
			if err != nil {
				fmt.Printf("%s", err.Error())
			} else {
				fmt.Println()
			}
			ret = false
		}
	}
	for _, item := range node.Content {
		if !processNode(item) {
			ret = false
		}
	}
	return ret
}
