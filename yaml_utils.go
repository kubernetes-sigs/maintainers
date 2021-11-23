package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml3 "gopkg.in/yaml.v3"
)

func removeUserFromOWNERS(path string, users []string) error {
	fmt.Printf("Fixing up %s\n", path)
	for _, user := range users {
		sourceYaml, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		rootNode := yaml3.Node{}
		err = yaml3.Unmarshal(sourceYaml, &rootNode)
		if err != nil {
			return err
		}

		switchToEmeritus(&rootNode, user)

		writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		encoder := yaml3.NewEncoder(writer)
		encoder.SetIndent(2)
		err = encoder.Encode(&rootNode)
		if err != nil {
			return err
		}
		err = encoder.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func switchToEmeritus(rootNode *yaml3.Node, user string) {
	// find mapping node
	mappingNode := fetchMappingNode(rootNode)
	if mappingNode == nil {
		return
	}

	// cleanup user from approvers and reviewers
	if !removeUserFromApproversAndReviewers(mappingNode, user) {
		return
	}

	// add user to emeritus list, create things we need if they are not there already
	emeritusSeqNode := fetchEmeritusSequenceNode(mappingNode)

	// add if not already present
	addUserToEmeritusList(emeritusSeqNode, user)
}

func addUserToEmeritusList(emeritusSeqNode *yaml3.Node, user string) {
	foundInEmeritusList := false
	for _, item := range emeritusSeqNode.Content {
		if item.Kind == yaml3.ScalarNode && strings.ToLower(item.Value) == strings.ToLower(user) {
			foundInEmeritusList = true
		}
	}
	if !foundInEmeritusList {
		node := yaml3.Node{
			Kind:  yaml3.ScalarNode,
			Tag:   "!!str",
			Value: user,
		}
		emeritusSeqNode.Content = append(emeritusSeqNode.Content, &node)
	}
}

func fetchEmeritusSequenceNode(mappingNode *yaml3.Node) *yaml3.Node {
	var emeritusScalar, emeritusSeqNode *yaml3.Node
	for i := 0; i < len(mappingNode.Content); i++ {
		node := mappingNode.Content[i]
		if node.Kind == yaml3.ScalarNode && (node.Value == "emeritus_approvers") {
			emeritusScalar = node
			j := i + 1
			if j < len(mappingNode.Content) && mappingNode.Content[j].Kind == yaml3.SequenceNode {
				emeritusSeqNode = mappingNode.Content[j]
			}
		}
	}
	if emeritusScalar == nil {
		emeritusScalar = &yaml3.Node{
			Kind:  yaml3.ScalarNode,
			Tag:   "!!str",
			Value: "emeritus_approvers",
		}
		mappingNode.Content = append(mappingNode.Content, emeritusScalar)
	}
	if emeritusSeqNode == nil {
		// remove trailing null scalar
		lastNode := mappingNode.Content[len(mappingNode.Content)-1]
		if lastNode.Kind == yaml3.ScalarNode && lastNode.Tag == "!!null" {
			mappingNode.Content = mappingNode.Content[:len(mappingNode.Content)-1]
		}
		emeritusSeqNode = &yaml3.Node{
			Kind: yaml3.SequenceNode,
			Tag:  "!!seq",
		}
		mappingNode.Content = append(mappingNode.Content, emeritusSeqNode)
	}
	return emeritusSeqNode
}

func fetchMappingNode(rootNode *yaml3.Node) *yaml3.Node {
	var mappingNode *yaml3.Node
	for _, node := range rootNode.Content {
		if node.Kind == yaml3.MappingNode {
			mappingNode = node
			break
		}
	}
	if mappingNode != nil {
		// structure of the file is slightly different for OWNERS_ALIASES
		foundAliases := false
		for _, node := range mappingNode.Content {
			if node.Kind == yaml3.ScalarNode && node.Value == "aliases" {
				foundAliases = true
			}
		}
		if foundAliases {
			for _, node := range mappingNode.Content {
				if node.Kind == yaml3.MappingNode {
					return node
				}
			}
		}
	}
	return mappingNode
}

func removeUserFromApproversAndReviewers(mappingNode *yaml3.Node, user string) bool {
	foundInApproverList := false
	for i := 0; i < len(mappingNode.Content); i++ {
		node := mappingNode.Content[i]
		if node.Kind == yaml3.ScalarNode && node.Value != "emeritus_approvers" {
			// found the approvers or reviewer node
			j := i + 1
			if j < len(mappingNode.Content) && mappingNode.Content[j].Kind == yaml3.SequenceNode {
				seqNode := mappingNode.Content[j]
				// found the list of approvers or reviewers
				var newList []*yaml3.Node
				for _, item := range seqNode.Content {
					// skip the user we want to eliminate from the list
					if item.Kind == yaml3.ScalarNode && strings.ToLower(item.Value) == strings.ToLower(user) {
						if node.Value == "approvers" {
							foundInApproverList = true
						}
					} else {
						newList = append(newList, item)
					}
				}
				seqNode.Content = newList
			}
		}
	}
	return foundInApproverList
}

