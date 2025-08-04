# Environments Configuration

This page explains how to configure environments for your repositories using viaplay-cli. Environments allow you to define deployment targets (such as staging, production, etc.) and set up environment-specific policies in your GitHub repositories.

---

## What is an Environment?

An environment in GitHub is a named context for deployments, such as `staging`, `production`, or `preview`. Each environment can have its own protection rules, reviewers, and secrets/variables.

---

## How to Configure Environments

Environments are configured using JSON files placed in your team configuration directory:

```bash
mkdir -p ~/.config/viaplay/teams/<team>/envs
```

Each environment should have its own JSON file. For example, to define a `staging` environment:

### Example: `staging.json`

```json
{
  "name": "staging",
  "wait_timer": 0,
  "reviewers": [],
  "deployment_branch_policy": {
    "protected_branches": false,
    "custom_branch_policies": true
  }
}
```

- `name`: The name of the environment (required)
- `wait_timer`: (optional) Wait timer before deployment (in seconds)
- `reviewers`: (optional) List of GitHub usernames required to approve deployments
- `deployment_branch_policy`: (optional) Branch protection settings for deployments

---

## Applying Environments

When you create a project or repository with viaplay-cli and use the `--apply-envs` flag, all environments defined in your team's `envs/` directory will be created in the GitHub repository.

```bash
vip create project --name myservice --team myteam --apply-envs
```

---

## Best Practices

- Define at least `staging` and `production` environments for most services.
- Use reviewers and branch policies to protect critical environments.
- Store environment secrets and variables using the secrets configuration.

---

For more details, see the [Project & Repo Creation](project-creation.md) and [Secrets](secrets.md) documentation.
