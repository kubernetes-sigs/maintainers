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

package utils

type OwnersInfo struct {
	Filters           map[string]FiltersInfo `json:"filters,omitempty"`
	Approvers         []string               `json:"approvers,omitempty"`
	Reviewers         []string               `json:"reviewers,omitempty"`
	RequiredReviewers []string               `json:"required_reviewers,omitempty"`
	Labels            []string               `json:"labels,omitempty"`
	EmeritusApprovers []string               `json:"emeritus_approvers,omitempty"`
	Options           DirOptions             `json:"options,omitempty"`
}

type DirOptions struct {
	NoParentOwners bool `json:"no_parent_owners,omitempty"`
}

type FiltersInfo struct {
	Approvers         []string `json:"approvers,omitempty"`
	Reviewers         []string `json:"reviewers,omitempty"`
	Labels            []string `json:"labels,omitempty"`
	EmeritusApprovers []string `json:"emeritus_approvers,omitempty"`
	RequiredReviewers []string `json:"required_reviewers,omitempty"`
}

// Aliases defines groups of people to be used in OWNERS files
type Aliases struct {
	RepoAliases map[string][]string `json:"aliases,omitempty"`
}

type Frames struct {
	Schema map[string]interface{} `json:"schema,omitempty"`
	Data   Values                 `json:"data,omitempty"`
}

type Values struct {
	Items [][]interface{} `json:"values,omitempty"`
}

type Contribution struct {
	ID           string
	Alias        string
	ContribCount int
	CommentCount int
}
