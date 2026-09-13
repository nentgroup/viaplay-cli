# Environments Configuration

Environments let you define deployment targets (such as staging, production, etc.) and set up environment-specific policies in your GitHub repositories using `vip`.

For detailed information about GitHub's Environment API, see the [GitHub Environments API documentation](https://docs.github.com/en/rest/deployments/environments?apiVersion=2022-11-28).

---

## What is an Environment?

An environment in GitHub is a named context for deployments, such as `staging`, `production`, or `preview`. Each environment can have its own protection rules, reviewers, and secrets/variables.

---

## How to Configure Environments

Environments are configured using YAML files placed in your team configuration directory:

```bash
mkdir -p ~/.config/viaplay/teams/<team>/envs
```

Each environment should have its own YAML file. For example, to define a `staging` environment:

### Example: `staging.yaml`

```yaml
name: staging
wait_timer: 0
reviewers:
  - type: Team
    id: {{.Org.TeamID}}
deployment_branch_policy:
  protected_branches: false
  custom_branch_policies: true
  branch_patterns:
    - name: main              # This represents a DeploymentBranchPolicyRequest
      type: branch            # Values could be "branch" or "tag"
    - name: "release/*"       # Another pattern example
      type: branch
    - name: "*"
      type: tag

```

- `name`: The name of the environment (required)
- `wait_timer`: (optional) Wait timer before deployment (in seconds)
- `reviewers`: (optional) List of GitHub usernames required to approve deployments
- `deployment_branch_policy`: (optional) Branch protection settings for deployments

---

## Applying Environments

When you create a project or repository with `vip` and use the `--apply-envs` flag, all environments defined in your team's `envs/` directory will be created in the GitHub repository.

```bash
vip project create myservice --team myteam --apply-envs
```

---

## Best Practices

- Define at least `staging` and `production` environments for most services.
- Use reviewers and branch policies to protect critical environments.
- Store environment secrets and variables using the secrets configuration.

---

For more details, see the [Project & Repo Creation](project-creation.md) and [Secrets](secrets.md) documentation.
