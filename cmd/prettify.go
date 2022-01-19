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
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/dims/maintainers/pkg/utils"
)

var indent int
var sigsyaml bool

func init() {
	prettifyCmd.Flags().IntVar(&indent, "indent", 2, "default indentation")
	prettifyCmd.Flags().BoolVar(&sigsyaml, "include-sigs-yaml", false, "indent sigs.yaml as well")
}

// exportCmd represents the export command
var prettifyCmd = &cobra.Command{
	Use:   "prettify",
	Short: "ensure all OWNERS related files are valid yaml and look the same",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		files, err := utils.GetOwnerFiles(pwd)
		if err != nil {
			panic(err)
		}

		aliasPath, err := utils.GetOwnersAliasesFile(pwd)
		if err == nil && len(aliasPath) > 0 {
			files = append(files, aliasPath)
		}

		if sigsyaml {
			sigsYamlPath, err := utils.GetSigsYamlFile(pwd)
			if err == nil && len(sigsYamlPath) > 0 {
				files = append(files, sigsYamlPath)
			}
		}

		for _, path := range files {
			sourceYaml, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			rootNode, err := fetchYaml(sourceYaml)
			if err != nil {
				panic(err)
			}
			writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				panic(err)
			}
			err = streamYaml(writer, indent, rootNode)
			if err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(prettifyCmd)
}

func fetchYaml(sourceYaml []byte) (*yaml.Node, error) {
	rootNode := yaml.Node{}
	err := yaml.Unmarshal(sourceYaml, &rootNode)
	if err != nil {
		return nil, err
	}
	return &rootNode, nil
}

func streamYaml(writer io.Writer, indent int, in *yaml.Node) error {
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(indent)
	err := encoder.Encode(in)
	if err != nil {
		return err
	}
	return encoder.Close()
}
