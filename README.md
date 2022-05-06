# Maintainers

Parse OWNERS, OWNER_ALIASES in kubernetes/kubernetes then pull information from devstats:
https://k8s.devstats.cncf.io/d/13/developer-activity-counts-by-repository-group?orgId=1&var-period_name=Last%20year&var-metric=contributions&var-repogroup_name=All&var-repo_name=kubernetes%2Fkubernetes&var-country_name=All

and print the github id, pr comment count and devstats contribution count as well.

## Installation

Run `make` to install `maintainers`. Currently `make` will run the `hack/install-maintainers.sh` script.

## Usage

Unless there is a flag to specify the context (e.g., --repository-github, --yaml-file,
--kubernetes-directory), the tool uses the directory in which it is run to provide the context
for the results. 

```bash
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

```bash
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

```bash
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

You can also validate/check if all the urls in a file are correct using `check-urls`
```bash
[dims@dims-m1 11:31] ~/go/src/k8s.io/community ⟩ ~/go/src/github.com/dims/maintainers/maintainers help check-urls
ensure all the urls in yaml file are still valid

Usage:
maintainers check-urls [flags]

Flags:
-h, --help               help for check-urls
--yaml-file string   validate urls in this yaml file (default "sigs.yaml")
```

The new `audit` command is helpful to kubernetes chairs and leads as it vets the sigs.yaml thoroughly.

Notes:
- ensure kubernetes/community and kubernetes/kubernetes is checked out under $GOPATH/src/k8s.io
- install the tool using `go install github.com/dims/maintainers@v0.2.0` (where v0.2.0 is latest tag as of right now, please check for newer tags)
- change directory to `$GOPATH/src/k8s.io/community` directory and run `maintainers audit sig-auth` (or your own sig) and review the output.

```bash
[dims@dims-m1-6163 15:28] ~/go/src/k8s.io/community ⟩ maintainers help audit
ensure OWNERS, OWNERS_ALIASES and sigs.yaml have the correct data structure

Usage:
  maintainers audit [name|all]... [flags]

Flags:
  -h, --help                          help for audit
      --kubernetes-directory string   path to kubernetes directory (default "/Users/dims/go/src/k8s.io/kubernetes")
```

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](http://slack.k8s.io/)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-dev)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[Creative Commons 4.0]: https://git.k8s.io/website/LICENSE
