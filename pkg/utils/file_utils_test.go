package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func utilJoinPath(t *testing.T, path string) (string, string) {
	utilsDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("error while get root dir %v", err)
	}
	pkgDir := filepath.Dir(utilsDir)
	root := filepath.Dir(pkgDir)
	testDir := filepath.Join(root, "testdata")
	file := filepath.Join(testDir, path)
	return file, testDir
}

func TestGetOwnersAliases(t *testing.T) {
	file, _ := utilJoinPath(t, "OWNERS_ALIASES")
	testCases := []struct {
		desc          string
		expectedState Aliases
	}{
		{
			desc: "using test OWNERS_ALIASES",
			expectedState: Aliases{
				RepoAliases: map[string][]string{
					"sig1-leads": {"lead1", "dude2", "guy3"},
					"sig2-leads": {"LEAD1", "DUDE2", "GUY3"},
					"wg1-leads":  {"lead@", "guy$"},
					"provider-1": {"lead1", "dude2"},
					"provider-2": {"DUDE2", "GUY3", "guy$"},
				},
			},
		},
	}
	for _, testCase := range testCases {
		result, err := GetOwnerAliases(file)
		if err != nil {
			t.Fatalf("error while get OWNERS_ALIASES for case %s: %v", testCase.desc, err)
		}
		if !reflect.DeepEqual(testCase.expectedState, *result) {
			t.Errorf("unexpected list of OWNERS Files for '%s', expected: %#v, got: %#v",
				testCase.desc,
				testCase.expectedState,
				*result)
		}
	}
}

func TestGetSigsYaml(t *testing.T) {
	file, _ := utilJoinPath(t, "sigs.yaml")
	testCases := []struct {
		desc          string
		expectedState Context
	}{
		{
			desc: "using test sigs.yaml",
			expectedState: Context{
				Sigs: []Group{
					{
						Dir:              "sig-api-machinery",
						Name:             "API Machinery",
						MissionStatement: "Covers all aspects of API server, API registration and discovery, generic API CRUD semantics, admission control, encoding/decoding, conversion, defaulting, persistence layer (etcd), OpenAPI, CustomResourceDefinition, garbage collection, and client libraries.\n",
						CharterLink:      "charter.md",
						StakeholderSIGs:  []string(nil),
						Label:            "api-machinery",
						Leadership: LeadershipGroup{
							Chairs: []Person{
								{
									GitHub:  "deads2k",
									Name:    "David Eads",
									Company: "Red Hat",
								},
							},
							TechnicalLeads: []Person{
								{
									GitHub:  "deads2k",
									Name:    "David Eads",
									Company: "Red Hat",
								},
							},
							EmeritusLeads: []Person(nil),
						},
						Meetings: []Meeting{
							{
								Description:   "Regular SIG Meeting",
								Day:           "Wednesday",
								Time:          "11:00",
								TZ:            "PT (Pacific Time)",
								Frequency:     "biweekly",
								URL:           "https://zoom.us/my/apimachinery",
								ArchiveURL:    "https://goo.gl/0lbiM9",
								RecordingsURL: "https://www.youtube.com/playlist?list=PL69nYSiGNLP21oW3hbLyjjj4XhrwKxH2R",
							},
						},
						Contact: Contact{
							Slack:              "sig-api-machinery",
							MailingList:        "https://groups.google.com/forum/#!forum/kubernetes-sig-api-machinery",
							PrivateMailingList: "",
							GithubTeams: []GithubTeam{
								{
									Name:        "sig-api-machinery-api-reviews",
									Description: "API Changes and Reviews (API Machinery APIs, NOT all APIs)",
								},
							},
							Liaison: &Person{GitHub: "dims", Name: "Davanum Srinivas"},
						},
						Subprojects: []Subproject{
							{
								Name:        "component-base",
								Description: "",
								Contact:     (*Contact)(nil),
								Owners: []string{
									"https://raw.githubusercontent.com/kubernetes-sigs/legacyflag/master/OWNERS",
								},
								Meetings: []Meeting(nil),
							},
						},
					},
				},
				WorkingGroups: []Group(nil),
				UserGroups:    []Group(nil),
				Committees:    []Group(nil),
			},
		},
	}
	for _, testCase := range testCases {
		result, err := GetSigsYaml(file)
		if err != nil {
			t.Fatalf("error while get OWNERS_ALIASES for case %v", err)
		}
		if !reflect.DeepEqual(testCase.expectedState, *result) {
			t.Errorf("unexpected list of OWNERS_ALIASES for '%s', expected: %#v, got: %#v",
				testCase.desc,
				testCase.expectedState,
				*result)
		}
	}
}
func TestGetOwnersInfo(t *testing.T) {
	file, _ := utilJoinPath(t, "OWNERS")
	testCases := []struct {
		desc          string
		expectedState OwnersInfo
	}{
		{
			desc: "using test OWNERS file",
			expectedState: OwnersInfo{
				Filters: map[string]FiltersInfo(nil),
				Approvers: []string{
					"dims", "MadhavJivrajani", "mrbobbytables", "nikhita", "palnabarun",
				},
				Reviewers: []string{
					"dims", "MadhavJivrajani", "mrbobbytables", "nikhita", "palnabarun",
				},
				RequiredReviewers: []string(nil),
				Labels:            []string(nil),
				EmeritusApprovers: []string(nil),
				Options: DirOptions{
					NoParentOwners: false,
				},
			},
		},
	}
	for _, testCase := range testCases {
		result, err := GetOwnersInfo(file)
		if err != nil {
			t.Fatalf("error while get OWNERS for case %v", err)
		}
		if !reflect.DeepEqual(testCase.expectedState, *result) {
			t.Errorf("unexpected list of OWNERS_ALIASES for '%s', expected: %#v, got: %#v",
				testCase.desc,
				testCase.expectedState,
				*result)
		}
	}
}

func TestGetOwnersInfoFromBytes(t *testing.T) {
	file, _ := utilJoinPath(t, "OWNERS")
	testCases := []struct {
		desc          string
		expectedState OwnersInfo
	}{
		{
			desc: "using test OWNERS",
			expectedState: OwnersInfo{
				Filters: map[string]FiltersInfo(nil),
				Approvers: []string{
					"dims", "MadhavJivrajani", "mrbobbytables", "nikhita", "palnabarun",
				},
				Reviewers: []string{
					"dims", "MadhavJivrajani", "mrbobbytables", "nikhita", "palnabarun",
				},
				RequiredReviewers: []string(nil),
				Labels:            []string(nil),
				EmeritusApprovers: []string(nil),
				Options: DirOptions{
					NoParentOwners: false,
				},
			},
		},
	}
	for _, testCase := range testCases {
		filename, _ := filepath.Abs(file)
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("error while get bytes for case %v", err)
		}
		result, err := GetOwnersInfoFromBytes(bytes)
		if err != nil {
			t.Fatalf("error while get OWNERS for case %v", err)
		}
		if !reflect.DeepEqual(testCase.expectedState, *result) {
			t.Errorf("unexpected list of OWNERS_ALIASES for '%s', expected: %#v, got: %#v",
				testCase.desc,
				testCase.expectedState,
				*result)
		}
	}
}

func TestGetOwnersAliasesFile(t *testing.T) {
	file, testDir := utilJoinPath(t, "OWNERS_ALIASES")
	testCases := []struct {
		desc          string
		expectedState string
	}{
		{
			desc:          "using test OWNERS_ALIASES",
			expectedState: file,
		},
	}
	for _, testCase := range testCases {
		result, err := GetOwnersAliasesFile(testDir)
		if err != nil {
			t.Fatalf("error while get OWNERS_ALIASES file for case %v", err)
		}
		if !(testCase.expectedState == result) {
			t.Errorf("unexpected OWNERS_ALIASES file for '%s', expected: %#v, got: %#v",
				testCase.desc,
				testCase.expectedState,
				result)
		}
	}
}

func TestGetOwnerFiles(t *testing.T) {
	file2, testDir := utilJoinPath(t, "OWNERS")
	root := filepath.Dir(testDir)
	file1 := filepath.Join(root, "OWNERS")
	testCases := []struct {
		desc          string
		expectedState []string
	}{
		{
			desc: "get OWNERS files",
			expectedState: []string{
				file1,
				file2,
			},
		},
	}
	for _, testCase := range testCases {
		result, err := GetOwnerFiles(root)
		if err != nil {
			t.Fatalf("error while get OWNERS files for case %v", err)
		}
		if !reflect.DeepEqual(testCase.expectedState, result) {
			t.Errorf("unexpected list of OWNERS Files for '%s', expected: %#v, got: %#v",
				testCase.desc,
				testCase.expectedState,
				result)
		}
	}
}

func TestSigsYamlFile(t *testing.T) {
	file, testDir := utilJoinPath(t, "sigs.yaml")
	testCases := []struct {
		desc          string
		expectedState string
	}{
		{
			desc:          "using test sigs.yaml",
			expectedState: file,
		},
	}
	for _, testCase := range testCases {
		result, err := GetSigsYamlFile(testDir)
		if err != nil {
			t.Fatalf("error while getting sigs.yaml file for case %v", err)
		}
		if !(testCase.expectedState == result) {
			t.Errorf("unexpected sigs.yaml file for '%s', expected: %#v, got: %#v",
				testCase.desc,
				testCase.expectedState,
				result)
		}
	}
}
