package main

type OwnersInfo struct {
	Filters   map[string]FiltersInfo `json:"filters,omitempty"`
	Approvers []string               `json:"approvers,omitempty"`
	Reviewers []string               `json:"reviewers,omitempty"`
}

type FiltersInfo struct {
	Approvers []string `json:"approvers,omitempty"`
	Reviewers []string `json:"reviewers,omitempty"`
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
	alias        string
	ContribCount int
	CommentCount int
}
