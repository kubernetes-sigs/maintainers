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

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type OwnersInfo struct {
	Filters           map[string]FiltersInfo `json:"filters,omitempty"`
	Approvers         []string               `json:"approvers,omitempty"`
	Reviewers         []string               `json:"reviewers,omitempty"`
	RequiredReviewers []string               `json:"required_reviewers,omitempty"`
	Labels            []string               `json:"labels,omitempty"`
	EmeritusApprovers []string               `json:"emeritus_approvers,omitempty"`
	EmeritusReviewers []string               `json:"emeritus_reviewers,omitempty"`
	Options           DirOptions             `json:"options,omitempty"`
}

func (o *OwnersInfo) EmeritusApproversCount() int {
	count := len(o.EmeritusApprovers)
	for _, f := range o.Filters {
		count += len(f.EmeritusApprovers)
	}

	return count
}

func (o *OwnersInfo) EmeritusReviewersCount() int {
	count := len(o.EmeritusReviewers)
	for _, f := range o.Filters {
		count += len(f.EmeritusReviewers)
	}

	return count
}

type DirOptions struct {
	NoParentOwners bool `json:"no_parent_owners,omitempty"`
}

type FiltersInfo struct {
	Approvers         []string `json:"approvers,omitempty"`
	Reviewers         []string `json:"reviewers,omitempty"`
	Labels            []string `json:"labels,omitempty"`
	EmeritusApprovers []string `json:"emeritus_approvers,omitempty"`
	EmeritusReviewers []string `json:"emeritus_reviewers,omitempty"`
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

// PrefixToGroupMap returns a map of prefix to groups, useful for iteration over all groups
func (c *Context) PrefixToGroupMap() map[string][]Group {
	return map[string][]Group{
		"sig":       c.Sigs,
		"wg":        c.WorkingGroups,
		"ug":        c.UserGroups,
		"committee": c.Committees,
	}
}

// DirName returns the directory that a group's documentation will be
// generated into. It is composed of a prefix (sig for SIGs and wg for WGs),
// and a formatted version of the group's name (in kebab case).
func (g *Group) DirName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(strings.Replace(g.Name, " ", "-", -1)))
}

// LabelName returns the expected label for a given group
func (g *Group) LabelName(prefix string) string {
	return strings.Replace(g.DirName(prefix), fmt.Sprintf("%s-", prefix), "", 1)
}

func GroupIndex(groups []Group, predicate func(Group) bool) int {
	for i, group := range groups {
		if predicate(group) {
			return i
		}
	}
	return -1
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

// PrefixToPersonMap returns a map of prefix to persons, useful for iteration over all persons
func (g *LeadershipGroup) PrefixToPersonMap() map[string][]Person {
	return map[string][]Person{
		"chair":         g.Chairs,
		"tech_lead":     g.TechnicalLeads,
		"emeritus_lead": g.EmeritusLeads,
	}
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

// EmeritusCounts holds mappings of path of an OWNERS file
// that has emeritus_{approvers,reviewers} to how many of
// them are there.
type EmeritusCounts struct {
	ReviewerCounts map[string]int
	ApproverCounts map[string]int
}

func NewEmeritusCounts() *EmeritusCounts {
	return &EmeritusCounts{
		ReviewerCounts: make(map[string]int),
		ApproverCounts: make(map[string]int),
	}
}

// EmeritusDiff captures the values calculated as the difference
// between two EmeritusCounts along with some additional info.
type EmeritusDiffFields struct {
	AddedCount            int
	RemovedCount          int
	OwnersFilesWhereAdded int
	OwnersFilesWhereDel   int
	AvgAddPerFile         float64
	AvgDelPerFile         float64
}

func (d EmeritusDiffFields) PrettyPrint() {
	fmt.Println("Number of OWNERS files where additions were done:", d.OwnersFilesWhereAdded)
	fmt.Println("Number of OWNERS files where deletions were done:", d.OwnersFilesWhereDel)
	fmt.Println("Total number added:", d.AddedCount)
	fmt.Println("Total number deleted:", d.RemovedCount)
	fmt.Println("Avg number added per OWNERS file:", d.AvgAddPerFile)
	fmt.Println("Avg number deleted per OWNERS file:", d.AvgDelPerFile)
}

type EmeritusDiff struct {
	Reviewers EmeritusDiffFields
	Approvers EmeritusDiffFields
}

func CalculateEmeritusDiff(from, to *EmeritusCounts) EmeritusDiff {
	diff := EmeritusDiff{}

	for path, count := range from.ReviewerCounts {
		if countTo, ok := to.ReviewerCounts[path]; ok {
			if countTo == count {
				continue
			}
			if countTo > count {
				diff.Reviewers.OwnersFilesWhereAdded++
				diff.Reviewers.AddedCount += (countTo - count)
			} else {
				diff.Reviewers.OwnersFilesWhereDel++
				diff.Reviewers.RemovedCount += (count - countTo)
			}
		}
	}

	for path, count := range from.ApproverCounts {
		if countTo, ok := to.ApproverCounts[path]; ok {
			if countTo == count {
				continue
			}
			if countTo > count {
				diff.Approvers.OwnersFilesWhereAdded++
				diff.Approvers.AddedCount += (countTo - count)
			} else {
				diff.Approvers.OwnersFilesWhereDel++
				diff.Approvers.RemovedCount += (count - countTo)
			}
		}
	}

	if diff.Reviewers.OwnersFilesWhereAdded != 0 {
		diff.Reviewers.AvgAddPerFile = float64(diff.Reviewers.AddedCount) / float64(diff.Reviewers.OwnersFilesWhereAdded)
	}
	if diff.Reviewers.OwnersFilesWhereDel != 0 {
		diff.Reviewers.AvgDelPerFile = float64(diff.Reviewers.RemovedCount) / float64(diff.Reviewers.OwnersFilesWhereDel)
	}

	if diff.Approvers.OwnersFilesWhereAdded != 0 {
		diff.Approvers.AvgAddPerFile = float64(diff.Approvers.AddedCount) / float64(diff.Approvers.OwnersFilesWhereAdded)
	}
	if diff.Approvers.OwnersFilesWhereDel != 0 {
		diff.Approvers.AvgDelPerFile = float64(diff.Approvers.RemovedCount) / float64(diff.Approvers.OwnersFilesWhereDel)
	}

	return diff
}
