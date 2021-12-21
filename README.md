Parse OWNERS, OWNER_ALIASES in kubernetes/kubernetes then pull information from devstats:
https://k8s.devstats.cncf.io/d/13/developer-activity-counts-by-repository-group?orgId=1&var-period_name=Last%20year&var-metric=contributions&var-repogroup_name=All&var-repo_name=kubernetes%2Fkubernetes&var-country_name=All

and print the github id, pr comment count and devstats contribution count as well.

```
[dims@dims-m1 17:41] ~/go/src/k8s.io/k8s.io ⟩ ../maintainers/maintainers --help
Usage of ../maintainers/maintainers:
      --dryrun                       do not modify any files (default true)
      --exclude strings              do not prune these comma-separated list of users from OWNERS
      --export                       export contents of all owners related files as output.csv
      --include strings              add these comma-separated list of users to prune from OWNERS
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
- use the `--export` to generate a CSV file with all the info in OWNERS, each line will
  have the github id, alias and the name of the file