# GitHub Actions Workflows

This directory contains CI/CD workflows for the notimulti notification server.

## Workflows

### `docker.yml` - Docker Build & Publish

Automated workflow for building, testing, and publishing Docker images.

**Triggers:**
- Push to `main` branch → Build and push with `latest` tag
- Push version tags (`v*.*.*`) → Build and push with semantic version tags
- Pull requests to `main` → Build only (no push)

**Workflow Stages:**

1. **Backend Tests** (`test-backend`)
   - Setup Go 1.21
   - Run all backend tests with race detector
   - Generate coverage report
   - **Enforce 80% coverage threshold** (NFR-018)
   - Upload coverage artifact

2. **Frontend Tests** (`test-frontend`)
   - Setup Node.js 18
   - Install dependencies with `npm ci`
   - Run test suite
   - Upload test results

3. **Build & Push** (`build-push`)
   - Only runs if both test jobs pass ✅
   - Build multi-architecture Docker image (amd64, arm64)
   - Push to Docker Hub (except for PRs)
   - Scan for vulnerabilities with Trivy
   - Apply semantic versioning tags

## Required GitHub Secrets

Configure these secrets in your repository settings (`Settings` → `Secrets and variables` → `Actions`):

### Docker Hub Credentials

| Secret | Description | How to Get |
|--------|-------------|------------|
| `DOCKERHUB_USERNAME` | Your Docker Hub username | Your Docker Hub account username |
| `DOCKERHUB_TOKEN` | Docker Hub access token | [Create token](https://hub.docker.com/settings/security) |

**Creating Docker Hub Access Token:**
1. Log in to [Docker Hub](https://hub.docker.com)
2. Go to Account Settings → Security
3. Click "New Access Token"
4. Name: `github-actions-notimulti`
5. Permissions: Read, Write, Delete
6. Copy the token (you won't see it again!)

### Optional Secrets

| Secret | Description | Required? |
|--------|-------------|-----------|
| `CODECOV_TOKEN` | Codecov.io upload token | No (coverage upload will be skipped) |

## Tagging Strategy

The workflow automatically creates Docker image tags based on the trigger:

### Version Tags (Push tag `v1.2.3`)
```bash
git tag v1.2.3
git push origin v1.2.3
```

Creates images:
- `notimulti:1.2.3` (full version)
- `notimulti:1.2` (minor version)
- `notimulti:1` (major version)
- `notimulti:latest` (if on main branch)

### Main Branch Push
```bash
git push origin main
```

Creates images:
- `notimulti:latest`
- `notimulti:sha-abc1234` (commit SHA)

### Pull Request
```bash
# Create PR from feature branch
```

Creates images (not pushed):
- `notimulti:pr-42` (PR number)
- `notimulti:sha-abc1234` (commit SHA)

## Multi-Architecture Support

Images are built for:
- **linux/amd64** - Intel/AMD 64-bit (most common)
- **linux/arm64** - ARM 64-bit (Apple Silicon, AWS Graviton, Raspberry Pi 4+)

Pull the image on any supported platform:
```bash
docker pull developertyrone/notimulti:latest
# Automatically pulls correct architecture
```

## Security Scanning

Every build is scanned with [Trivy](https://github.com/aquasecurity/trivy) for vulnerabilities.

**Scan Results:**
- **CRITICAL/HIGH** vulnerabilities → Build fails ❌
- **MEDIUM/LOW** vulnerabilities → Build continues with warning ⚠️

View scan results in:
- Workflow run logs → "Run Trivy vulnerability scanner" step
- GitHub Security tab → Code scanning alerts

## Workflow Execution

### Manual Trigger

Trigger workflow manually from GitHub UI:
1. Go to `Actions` tab
2. Select "Docker Build & Publish"
3. Click "Run workflow"
4. Select branch
5. Click "Run workflow"

### Automatic Trigger Examples

**Release a new version:**
```bash
# Create and push version tag
git tag v1.0.0
git push origin v1.0.0

# Workflow runs automatically
# Pushes to: notimulti:1.0.0, notimulti:1.0, notimulti:1, notimulti:latest
```

**Deploy latest changes:**
```bash
# Merge PR to main
git checkout main
git merge feature-branch
git push origin main

# Workflow runs automatically
# Pushes to: notimulti:latest, notimulti:sha-abc1234
```

**Test changes in PR:**
```bash
# Create PR from feature branch
gh pr create --title "Add new feature"

# Workflow runs automatically
# Builds but doesn't push (test only)
```

## Layer Caching

The workflow uses GitHub Actions cache to speed up builds:

**First build:** ~5-8 minutes (full build)  
**Subsequent builds:** ~2-4 minutes (cached layers)

Cache is automatically managed by GitHub (no configuration needed).

## Troubleshooting

### Build Fails: "Coverage below 80%"

**Problem:** Backend test coverage is below required threshold.

**Solution:**
```bash
# Check coverage locally
cd backend
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

# Add tests to increase coverage
# Re-run and verify ≥80%
```

### Build Fails: "Docker Hub login failed"

**Problem:** Docker Hub credentials are missing or incorrect.

**Solution:**
1. Verify secrets are configured:
   - `DOCKERHUB_USERNAME` exists
   - `DOCKERHUB_TOKEN` exists
2. Check token hasn't expired
3. Regenerate token if needed
4. Update secret with new token

### Build Fails: "Trivy scan found CRITICAL vulnerabilities"

**Problem:** Base image or dependencies have known security issues.

**Solution:**
1. Check scan results in workflow logs
2. Update base image in Dockerfile:
   ```dockerfile
   FROM alpine:3.18  →  FROM alpine:3.19
   ```
3. Update Go dependencies:
   ```bash
   cd backend
   go get -u ./...
   go mod tidy
   ```
4. Rebuild and verify

### Build Fails: "QEMU setup failed"

**Problem:** Multi-architecture build setup issue.

**Solution:**
- Usually transient GitHub Actions issue
- Re-run the workflow
- If persists, check GitHub Actions status page

### Build Succeeds but Image Not on Docker Hub

**Problem:** Image built but not pushed.

**Possible Causes:**
1. **Pull Request** - Images are only built, not pushed for PRs (by design)
2. **Branch** - Only `main` branch and version tags push images
3. **Secrets** - Docker Hub credentials missing/incorrect

**Solution:**
- For PRs: Merge to main to push
- For other branches: Push to main or create version tag
- Check secrets are configured correctly

### Frontend Tests Fail: "npm ci failed"

**Problem:** Node.js dependency installation issue.

**Solution:**
1. Check `package-lock.json` is committed
2. Verify Node.js version compatibility:
   ```json
   "engines": {
     "node": ">=18.0.0"
   }
   ```
3. Run locally to reproduce:
   ```bash
   cd frontend
   rm -rf node_modules package-lock.json
   npm install
   npm test
   ```

### Backend Tests Fail: "race detector found data race"

**Problem:** Concurrent access to shared state without synchronization.

**Solution:**
1. Run locally with race detector:
   ```bash
   cd backend
   go test -race ./...
   ```
2. Fix data races (add mutexes, channels, or sync primitives)
3. Re-test until clean

### Workflow Doesn't Trigger

**Problem:** Push/tag created but workflow not running.

**Solution:**
1. Check workflow file syntax:
   ```bash
   # Use YAML linter
   yamllint .github/workflows/docker.yml
   ```
2. Verify workflow is enabled:
   - Go to Actions tab
   - Check if workflow is disabled
3. Check branch protection rules don't block workflow

## Performance Optimization

**Reduce build time:**
1. ✅ Layer caching enabled (automatic)
2. ✅ Use `npm ci` instead of `npm install` (deterministic, faster)
3. ✅ Dependency caching for Go and Node.js
4. ✅ Multi-stage Dockerfile (optimized layer ordering)

**Current Build Times:**
- Backend tests: ~1-2 minutes
- Frontend tests: ~1-2 minutes
- Docker build (cached): ~2-4 minutes
- Docker build (uncached): ~5-8 minutes
- **Total (with cache): ~4-8 minutes**

## Monitoring

### View Workflow Runs

1. Go to repository Actions tab
2. Select "Docker Build & Publish"
3. View run history with status indicators:
   - ✅ Green checkmark - Success
   - ❌ Red X - Failed
   - 🟡 Yellow dot - In progress
   - ⚪ Gray circle - Queued

### Check Coverage Reports

1. Go to workflow run
2. Expand "Backend Tests" job
3. Check "Check coverage threshold" step for percentage

### View Docker Hub Images

1. Visit https://hub.docker.com/r/developertyrone/notimulti
2. Check "Tags" tab for all published versions
3. View pull statistics and metadata

## Best Practices

1. **Always test locally before pushing:**
   ```bash
   cd backend && go test ./...
   cd frontend && npm test
   docker build -t notimulti:test .
   ```

2. **Use semantic versioning for releases:**
   ```bash
   # Major version (breaking changes)
   git tag v2.0.0
   
   # Minor version (new features)
   git tag v1.1.0
   
   # Patch version (bug fixes)
   git tag v1.0.1
   ```

3. **Keep workflows fast:**
   - Cache dependencies
   - Run tests in parallel
   - Use minimal base images

4. **Monitor security:**
   - Review Trivy scan results
   - Update dependencies regularly
   - Subscribe to GitHub Security Advisories

5. **Protect credentials:**
   - Never commit Docker Hub tokens
   - Use GitHub Secrets
   - Rotate tokens every 90 days

## Further Reading

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Trivy Security Scanner](https://github.com/aquasecurity/trivy)
- [Docker Hub](https://hub.docker.com)
- [Semantic Versioning](https://semver.org)
