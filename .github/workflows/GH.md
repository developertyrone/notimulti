# GitHub Actions secret checklist

The workflows under `.github/workflows/` expect the following repository secrets to be configured. Set them under **Repository Settings → Secrets and variables → Actions**.

| Secret | Required by | When it is used | Notes |
| --- | --- | --- | --- |
| `DOCKERHUB_USERNAME` | `docker.yml` (Build & Publish job) | Needed on pushes/tags to `main` when logging in with `docker/login-action` | Docker Hub account that owns the `notimulti` image namespace. Login step is skipped for pull requests, but the secret must exist for release builds.
| `DOCKERHUB_TOKEN` | `docker.yml` (Build & Publish job) | Paired with `DOCKERHUB_USERNAME` during `docker/login-action` | Create a Docker Hub access token or use the account password. Must have permission to push to the target repository.
| `CODECOV_TOKEN` | `docker.yml` (backend test job) | Used when uploading backend coverage to Codecov on non-PR runs | Optional for forks/PRs. Obtain the token from the Codecov project settings.

## Managing secrets with `gh`

Set secrets directly from the CLI (assumes you are in the repo root and authenticated with `gh auth login`):

```bash
gh secret set DOCKERHUB_USERNAME --body "<dockerhub username>"
gh secret set DOCKERHUB_TOKEN --body "<dockerhub token>"
gh secret set CODECOV_TOKEN --body "<codecov token>"
```

List all configured secrets to verify:

```bash
gh secret list
```

> Tip: if you eventually move the image to another registry, also update the `IMAGE_NAME` default (currently `${github.repository_owner}/notimulti`).