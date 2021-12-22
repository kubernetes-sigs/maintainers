Parse OWNERS, OWNER_ALIASES in kubernetes/kubernetes then pull information from devstats:
https://k8s.devstats.cncf.io/d/13/developer-activity-counts-by-repository-group?orgId=1&var-period_name=Last%20year&var-metric=contributions&var-repogroup_name=All&var-repo_name=kubernetes%2Fkubernetes&var-country_name=All

and print the github id, pr comment count and devstats contribution count as well.

```
[dims@dims-m1 20:58] ~/go/src/k8s.io/kubernetes ⟩ ../maintainers/maintainers help prune
Remove stale github ids from OWNERS and OWNERS_ALIASES

Usage:
  maintainers prune [flags]

Flags:
      --dryrun                       do not modify any files (default true)
      --exclude strings              do not prune these comma-separated list of users from OWNERS
  -h, --help                         help for prune
      --include strings              add these comma-separated list of users to prune from OWNERS
      --period-devstats string       one of "y" (year) "q" (quarter) "m" (month)  (default "y")
      --repository-devstats string   defaults to "kubernetes/kubernetes" repository (default "kubernetes/kubernetes")
      --repository-github string     defaults to "kubernetes/kubernetes" repository (default "kubernetes/kubernetes")
      --skip-devstats                skip devstat contributions count check
      --skip-github                  skip github PR count check
```

Notes:
- Use `--dryrun=true` to update all the files
- You can specify the repositories from where to fetch the contribution or PR 
  comments using `--repository-devstats` or `--repository-github`
- If you want to skip either the devstats check or the github check use the corresponding flag, either
  `--skip-devstats` or `--skip-github`
- Use `include` or `exclude` to tune who gets removed
- You can even add both the skips and use the `include` to remove specific users

```
[dims@dims-m1 20:59] ~/go/src/k8s.io/kubernetes ⟩ ../maintainers/maintainers help export
export contents of OWNERS and OWNERS_ALIASES as parsable csv file

Usage:
maintainers export [flags]

Flags:
-h, --help   help for export
```

Notes:
- use the `export` to generate a CSV file with all the info in OWNERS, each line will
  have the github id, alias and the name of the file

```
[dims@dims-m1 08:33] ~/go/src/k8s.io/community ⟩ ../maintainers/maintainers help prettify
ensure all OWNERS related files are valid yaml and look the same

Usage:
  maintainers prettify [flags]

Flags:
  -h, --help         help for prettify
      --indent int   default indentation (default 2)
```

ensure OWNERS files to have a consistent format (spaces, line breaks etc) for any automation to be built around
updating these files.