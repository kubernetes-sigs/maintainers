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
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/dims/maintainers/pkg/utils"
)

var kubernetesDirectory string

func getDefaultKubernetesDirectory() string {
	val, ok := os.LookupEnv("GOPATH")
	if !ok {
		fmt.Print("WARN: GOPATH not set")
		return ""
	}

	return path.Join(val, "src/k8s.io/kubernetes")
}

func init() {
	auditCmd.Flags().StringVar(&kubernetesDirectory, "kubernetes-directory", getDefaultKubernetesDirectory(), "path to kubernetes directory")
	rootCmd.AddCommand(auditCmd)
}

// auditCmd represents the audit command
var auditCmd = &cobra.Command{
	Use:   "audit [name|all]...",
	Args:  cobra.MinimumNArgs(1),
	Short: "ensure OWNERS, OWNERS_ALIASES and sigs.yaml have the correct data structure",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Running script : %s\n", time.Now().Format("01-02-2006 15:04:05"))
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		if _, err := os.Stat(kubernetesDirectory); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("please use --kubernetes-directory to set the path to the kubernetes directory. "+
				"%s does not exist\n", kubernetesDirectory)
			os.Exit(1)
		}

		sigsYamlPath, err := utils.GetSigsYamlFile(pwd)
		if err != nil {
			panic(fmt.Errorf("ERROR: unable to find sigs.yaml file: %w", err))
		}
		context, err := utils.GetSigsYaml(sigsYamlPath)
		if err != nil {
			panic(fmt.Errorf("ERROR: parsing file: %s - %w", sigsYamlPath, err))
		}

		if auditSpecifiedGroups(pwd, context, args) {
			auditGithubIDs(context)
			auditLocalOwnersFiles(context, args)
		}
		fmt.Printf("Done.\n")
	},
}

func auditLocalOwnersFiles(context *utils.Context, args []string) {
	fmt.Printf("\n>>>> Processing owners files\n")
	mapFilesToGroups := make(map[string]sets.String)
	var listOfGroups []string
	for _, groups := range context.PrefixToGroupMap() {
		for _, group := range groups {
			listOfGroups = append(listOfGroups, group.Dir)
			var files []string
			for _, subproject := range group.Subprojects {
				for _, owner := range subproject.Owners {
					if strings.Contains(owner, "/kubernetes/kubernetes/") {
						split := strings.SplitN(owner, "/", 7)
						filename := split[len(split)-1]
						files = append(files, filename)
						if val, ok := mapFilesToGroups[filename]; ok {
							val.Insert(group.Dir)
						} else {
							val := sets.String{}
							val.Insert(group.Dir)
							mapFilesToGroups[filename] = val
						}
					}
				}
			}
		}
	}
	sort.Strings(listOfGroups)
	files, err := utils.GetOwnerFiles(kubernetesDirectory)
	if err != nil {
		fmt.Printf("ERROR: unable to find kubernetes directory - %s\n", err)
		return
	}
	infoLog := sets.String{}
	for _, file := range files {
		likelyGroups := sets.String{}
		info, err := utils.GetOwnersInfo(file)
		if err != nil {
			fmt.Printf("ERROR: unable to read file %s - %s\n", file, err)
			continue
		}
		for _, label := range info.Labels {
			label = strings.ReplaceAll(label, "/", "-")
			for _, g := range listOfGroups {
				if strings.HasPrefix(label, g) {
					likelyGroups.Insert(g)
				}
			}
		}
		allOwners := []string{}
		allOwners = append(allOwners, info.Approvers...)
		allOwners = append(allOwners, info.Reviewers...)
		allOwners = append(allOwners, info.RequiredReviewers...)
		for _, item := range allOwners {
			for _, g := range listOfGroups {
				if strings.HasPrefix(item, g) {
					likelyGroups.Insert(g)
				}
			}
		}
		subpath := strings.Replace(file, kubernetesDirectory, "", -1)[1:]
		candidates := likelyGroups.List()
		if val, ok := mapFilesToGroups[subpath]; ok {
			actualGroups := val.List()
			if len(candidates) != 0 {
				if !reflect.DeepEqual(actualGroups, candidates) {
					if groupNameInArgs(candidates, args) || groupNameInArgs(actualGroups, args) {
						infoLog.Insert(fmt.Sprintf("ERROR: file %s should be in %q based on labels/aliases but is in %q\n",
							subpath, candidates, actualGroups))
					}
				}
			}
		} else {
			if len(candidates) > 0 {
				if groupNameInArgs(candidates, args) {
					infoLog.Insert(fmt.Sprintf("WARNING: file %s should be in one of %q based on labels/aliases\n",
						subpath, candidates))
				}
			} else {
				infoLog.Insert(fmt.Sprintf("INFO: unable to classify %s\n", subpath))
			}
		}
	}
	for _, line := range infoLog.List() {
		fmt.Printf(line)
	}
}

func groupNameInArgs(groupNames []string, args []string) bool {
	for _, groupName := range groupNames {
		for _, name := range args {
			if name == "all" || groupName == name || strings.Contains(groupName, name) {
				return true
			}
		}
	}
	return false
}

func auditGithubIDs(context *utils.Context) {
	fmt.Printf("\n>>>> Processing github id(s)\n")
	people := make(map[string]utils.Person)
	for groupType, groups := range context.PrefixToGroupMap() {
		for _, group := range groups {
			for prefix, persons := range group.Leadership.PrefixToPersonMap() {
				for _, person := range persons {
					if val, ok := people[person.GitHub]; ok {
						if val.Name != person.Name || (prefix != "emeritus_lead" && val.Company != person.Company) {
							fmt.Printf("ERROR: %s/%s: %s: expected person: %v, got: %v\n", groupType, group.Dir, prefix, val, person)
						}
					} else if prefix != "emeritus_lead" {
						people[person.GitHub] = person
					}

					if prefix == "emeritus_lead" && person.Company != "" {
						fmt.Printf("ERROR: %s/%s: emeritus leads should not have company specified; company specified for: %s\n", groupType, group.Dir, person.Name)
					}
				}
			}
		}
	}
	// TODO: grab contribution stats to see who is active?
}

func auditSpecifiedGroups(pwd string, context *utils.Context, args []string) bool {
	found := false
	for _, name := range args {
		for groupType, groups := range context.PrefixToGroupMap() {
			for _, group := range groups {
				if name == "all" || strings.Contains(group.Name, name) || strings.Contains(group.Dir, name) {
					auditGroup(pwd, groupType, group, context)
					found = true
				}
			}
		}
		if !found {
			fmt.Printf("[%s] not found\n", name)
		}
	}
	return found
}

func auditGroup(pwd string, groupType string, group utils.Group, context *utils.Context) {
	if len(group.Dir) == 0 {
		fmt.Printf("WARNING: missing 'dir' key\n")
	}
	if len(group.Name) == 0 {
		fmt.Printf("WARNING: missing 'name' key\n")
	}
	fmt.Printf("\n>>>> Processing %s [%s/%s]\n", groupType, group.Dir, group.Name)

	expectedDir := group.DirName(groupType)
	if expectedDir != group.Dir {
		fmt.Printf("ERROR: expected dir: %s, got: %s\n", expectedDir, group.Dir)
	}
	expectedLabel := group.LabelName(groupType)
	if expectedLabel != group.Label {
		fmt.Printf("ERROR: expected label: %s, got: %s\n", expectedLabel, group.Label)
	}
	if groupType == "sig" {
		if len(group.MissionStatement) == 0 {
			fmt.Printf("ERROR: missing 'mission_statement' key\n")
		}
		if len(group.CharterLink) == 0 {
			fmt.Printf("ERROR: missing 'charter_link' key\n")
		} else {
			auditCharterLink(pwd, group)
		}
	}
	if groupType == "wg" {
		auditWorkingGroupStakeholders(groupType, group, context)
	}
	if len(group.Label) == 0 {
		fmt.Printf("WARNING: missing 'label' keys\n")
	}
	auditLeadership(group, groupType)
	if len(group.Meetings) == 0 {
		fmt.Printf("WARNING: missing 'meetings' key\n")
	}
	auditContact(&group.Contact)
	if groupType == "sig" {
		if len(group.Subprojects) == 0 {
			fmt.Printf("WARNING: missing 'subprojects' key\n")
		} else {
			auditSubProject(groupType, group)
		}
	}
	if groupType != "committee" && groupType != "sig" {
		if len(group.Subprojects) > 0 {
			fmt.Printf("ERROR: only sigs and committees can own code / have subprojects, found: %d subprojects\n", len(group.Subprojects))
		}
	}
}

func auditSubProject(groupType string, group utils.Group) {
	for _, subproject := range group.Subprojects {
		fmt.Printf("\n>>>> Processing subproject %s under %s\n", subproject.Name, group.Dir)
		if len(subproject.Name) == 0 {
			fmt.Printf("WARNING: missing 'name' key\n")
		}
		if len(subproject.Description) == 0 {
			fmt.Printf("WARNING: missing 'description' key\n")
		}
		if subproject.Contact == nil {
			fmt.Printf("WARNING: missing 'contact' key\n")
		} else {
			auditContact(subproject.Contact)
		}
		if len(subproject.Owners) == 0 {
			fmt.Printf("ERROR: missing 'owners' key\n")
		} else {
			auditOwnersFiles(groupType, group, subproject)
		}
		if len(subproject.Meetings) == 0 {
			fmt.Printf("WARNING: missing 'meetings' key\n")
		}
	}
}

const (
	regexRawGitHubURL = "https://raw.githubusercontent.com/(?P<org>[^/]+)/(?P<repo>[^/]+)/(?P<branch>[^/]+)/(?P<path>.*)"
	regexGitHubURL    = "https://github.com/(?P<org>[^/]+)/(?P<repo>[^/]+)/(blob|tree)/(?P<branch>[^/]+)/(?P<path>.*)"
)

var reRawGitHubURL, reGitHubURL *regexp.Regexp
var regexpOnce sync.Once

func auditOwnersFiles(groupType string, group utils.Group, subproject utils.Subproject) {
	regexpOnce.Do(func() {
		reRawGitHubURL = regexp.MustCompile(regexRawGitHubURL)
		reGitHubURL = regexp.MustCompile(regexGitHubURL)
	})
	fmt.Printf("\n>>>> Processing owners files for %s/%s\n", group.Dir, subproject.Name)
	if len(subproject.Owners) == 0 {
		fmt.Printf("ERROR: subproject %s has no owners\n", subproject.Name)
	}
	for _, url := range subproject.Owners {
		if !reRawGitHubURL.MatchString(url) && !reGitHubURL.MatchString(url) {
			fmt.Printf("ERROR: owner urls should match regexp %s, found: %s\n", regexRawGitHubURL, url)
			continue
		}
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("ERROR: unable to read owners file at %s url - %v\n", url, err)
			}
			info, err := utils.GetOwnersInfoFromBytes(bytes)
			if err != nil {
				fmt.Printf("ERROR: unable to parse owners file at %s url - %v\n", url, err)
			} else {
				if !strings.Contains(url, "kubernetes/kubernetes") {
					continue
				}
				auditOwnersInfo(groupType, group, info, url)
			}
		} else {
			fmt.Printf("WARNING: stale url in %s - %s - http status code = %d - %s\n",
				group.DirName(groupType), url, resp.StatusCode, err)
		}
	}
}

func auditOwnersInfo(groupType string, group utils.Group, info *utils.OwnersInfo, url string) {
	lookFor := group.DirName(groupType)
	if len(info.Labels) > 0 {
		if len(group.Label) > 0 {
			found := false
			for _, label := range info.Labels {
				if strings.HasSuffix(label, group.Label) {
					found = true
				}
			}
			if !found {
				fmt.Printf("WARNING: needs labels reflecting %s - %s\n", lookFor, url)
			}
		}
	} else {
		fmt.Printf("WARNING: needs labels reflecting %s - %s\n", lookFor, url)
	}
	allOwners := []string{}
	allOwners = append(allOwners, info.Approvers...)
	allOwners = append(allOwners, info.Reviewers...)
	allOwners = append(allOwners, info.RequiredReviewers...)
	found := false
	for _, item := range allOwners {
		if strings.Contains(item, lookFor) {
			found = true
		}
	}
	if !found {
		fmt.Printf("WARNING: needs an alias as approver/reviewer reflecting %s - %s\n", lookFor, url)
	}
}

func auditContact(contact *utils.Contact) {
	if len(contact.Slack) == 0 {
		fmt.Printf("WARNING: missing 'slack' in contact\n")
	}
	if len(contact.MailingList) == 0 {
		fmt.Printf("WARNING: missing 'mailing_list' in contact\n")
	}
	if len(contact.PrivateMailingList) == 0 {
		fmt.Printf("OPTIONAL: missing 'private_mailing_list' in contact\n")
	}
	if len(contact.GithubTeams) == 0 {
		fmt.Printf("OPTIONAL: missing 'teams' in contact\n")
	}
	if contact.Liaison != nil {
		auditPerson("contact/liaison", contact.Liaison)
	}
}

func auditCharterLink(pwd string, group utils.Group) {
	if strings.HasPrefix(group.CharterLink, "http") {
		client := &http.Client{}
		resp, err := client.Get(group.CharterLink)
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Printf("WARNING: unable to reach url for 'charter_link' - %s\n", group.CharterLink)
		}
	} else {
		charterPath := path.Join(pwd, group.Dir, group.CharterLink)
		if _, err := os.Stat(charterPath); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("WARNING: missing file for 'charter_link' - %s\n", charterPath)
		}
	}
}

func auditWorkingGroupStakeholders(groupType string, group utils.Group, context *utils.Context) {
	if groupType == "wg" {
		if len(group.StakeholderSIGs) == 0 {
			fmt.Printf("WARNING: missing 'stakeholder_sigs' key\n")
		} else {
			for _, stakeholder := range group.StakeholderSIGs {
				found := false
				for _, group := range context.Sigs {
					if group.Name == stakeholder {
						found = true
						break
					}
				}
				if !found {
					fmt.Printf("WARNING: stakeholder_sigs entry '%s' not found (typo?)\n", stakeholder)
				}
			}
		}
	} else {
		if len(group.StakeholderSIGs) > 0 {
			fmt.Printf("ERROR: only 'workinggroups' may have stakeholder_sigs ()\n")
		}
	}
}

func auditLeadership(group utils.Group, groupType string) {
	if len(group.Leadership.Chairs) == 0 {
		fmt.Printf("WARNING: missing 'chairs' key (in 'leadership' section)\n")
		if groupType == "sig" {
			if len(group.Leadership.Chairs) == 1 {
				fmt.Printf("WARNING: please consider adding more folks in as 'chairs' (in 'leadership' section)\n")
			}
		}
	}
	if len(group.Leadership.TechnicalLeads) == 0 {
		fmt.Printf("WARNING: missing 'tech_leads' key (in 'leadership' section)\n")
		if groupType == "sig" {
			fmt.Printf("WARNING: if chairs are serving as tech leads, please add them explicitly in 'tech_leads' key (in 'leadership' section)\n")
		}
	}
	var persons []utils.Person
	persons = append(persons, group.Leadership.Chairs...)
	persons = append(persons, group.Leadership.TechnicalLeads...)
	persons = append(persons, group.Leadership.EmeritusLeads...)
	for _, person := range persons {
		auditPerson("leadership", &person)
	}
}

func auditPerson(extra string, person *utils.Person) {
	if len(person.Name) == 0 {
		fmt.Printf("WARNING: missing 'name' key in %s\n", extra)
	}
	if len(person.GitHub) == 0 {
		fmt.Printf("WARNING: missing 'github' key in %s for %s\n", extra, person.Name)
	}
}
