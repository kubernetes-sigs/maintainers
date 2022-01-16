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

import "gopkg.in/yaml.v3"

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

// Context is the context for the sigs.yaml file.
type Context struct {
	Sigs          []Group
	WorkingGroups []Group
	UserGroups    []Group
	Committees    []Group
}

// Group represents either a Special Interest Group (SIG) or a Working Group (WG)
type Group struct {
	Dir              string
	Name             string
	MissionStatement FoldedString `yaml:"mission_statement,omitempty" json:"mission_statement,omitempty"`
	CharterLink      string       `yaml:"charter_link,omitempty" json:"charter_link,omitempty"`
	StakeholderSIGs  []string     `yaml:"stakeholder_sigs,omitempty" json:"stakeholder_sigs,omitempty"`
	Label            string
	Leadership       LeadershipGroup `yaml:"leadership" json:"leadership"`
	Meetings         []Meeting
	Contact          Contact
	Subprojects      []Subproject `yaml:",omitempty" json:",omitempty"`
}

// GithubTeam represents a specific Github Team.
type GithubTeam struct {
	Name        string
	Description string `yaml:",omitempty" json:",omitempty"`
}

// Subproject represents a specific subproject owned by the group
type Subproject struct {
	Name        string
	Description string   `yaml:",omitempty" json:",omitempty"`
	Contact     *Contact `yaml:",omitempty" json:",omitempty"`
	Owners      []string
	Meetings    []Meeting `yaml:",omitempty" json:",omitempty"`
}

// LeadershipGroup represents the different groups of leaders within a group
type LeadershipGroup struct {
	Chairs         []Person
	TechnicalLeads []Person `yaml:"tech_leads,omitempty" json:"tech_leads,omitempty"`
	EmeritusLeads  []Person `yaml:"emeritus_leads,omitempty" json:"emeritus_leads,omitempty"`
}

// Person represents an individual person holding a role in a group.
type Person struct {
	GitHub  string
	Name    string
	Company string `yaml:"company,omitempty" json:",omitempty"`
}

// Meeting represents a regular meeting for a group.
type Meeting struct {
	Description   string
	Day           string
	Time          string
	TZ            string
	Frequency     string
	URL           string `yaml:",omitempty" json:",omitempty"`
	ArchiveURL    string `yaml:"archive_url,omitempty" json:"archive_url,omitempty"`
	RecordingsURL string `yaml:"recordings_url,omitempty" json:"recordings_url,omitempty"`
}

// Contact represents the various contact points for a group.
type Contact struct {
	Slack              string       `yaml:",omitempty" json:",omitempty"`
	MailingList        string       `yaml:"mailing_list,omitempty" json:"mailing_list,omitempty"`
	PrivateMailingList string       `yaml:"private_mailing_list,omitempty" json:"private_mailing_list,omitempty"`
	GithubTeams        []GithubTeam `yaml:"teams,omitempty" json:"teams,omitempty"`
	Liaison            *Person      `yaml:"liaison,omitempty" json:"liaison,omitempty"`
}

// FoldedString is a string that will be serialized in FoldedStyle by go-yaml
type FoldedString string

// MarshalYAML customizes how FoldedStrings will be serialized by go-yaml
func (x FoldedString) MarshalYAML() (interface{}, error) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Style: yaml.FoldedStyle,
		Value: string(x),
	}, nil
}
