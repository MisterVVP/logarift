# Package Publishing Runbook

Logarift publishes runtime containers and Helm artifacts to GitHub Container Registry (GHCR) under `ghcr.io/mistervvp`. This makes the image repositories referenced by the Helm chart available before users install the chart.

## Published artifacts

The release workflow publishes these OCI container images:

| Component | Image |
| --- | --- |
| Backend API | `ghcr.io/mistervvp/logarift-api` |
| Frontend | `ghcr.io/mistervvp/logarift-frontend` |
| Math engine | `ghcr.io/mistervvp/logarift-math-engine` |
| Optional LLM adapter | `ghcr.io/mistervvp/logarift-llm-adapter` |

It also publishes the Helm chart as an OCI artifact at:

```text
ghcr.io/mistervvp/charts/logarift
```

For GitHub releases, the workflow uploads these release assets in addition to GitHub's automatically generated source archives:

- packaged Helm chart, for example `logarift-1.2.3.tgz`
- repository source tarball, for example `logarift-1.2.3-source.tar.gz`
- repository source ZIP archive, for example `logarift-1.2.3-source.zip`
- `SHA256SUMS` for the uploaded release assets

## Publish triggers

Packages are published by `.github/workflows/release-packages.yml` on:

- GitHub Release publication (`release.published`)
- pushes to `main`
- pushes to branches named `dev-*`
- manual `workflow_dispatch` runs

This supports continuously refreshed `latest` packages from `main`, release distribution, and branch-based integration testing before a release is cut.

## Tagging policy

Release publications use the Git tag as the version source. For a release tag such as `v1.2.3`, images and the Helm chart are published with version `1.2.3`. Non-prerelease publications also update `1.2` and `latest` image tags. The source chart defaults to `latest` image tags for local installs, and the release workflow rewrites packaged chart image tags to the release, `main`, or branch app version before linting and publishing.

Main branch publications update mutable `latest` image tags and immutable short-SHA image tags on every push. The Helm chart is published with the chart base version plus a run-number prerelease suffix, for example `0.1.0-main.42`, and its app version defaults to `latest` so installs target the refreshed main images.

Development branch publications use mutable branch image tags and immutable short-SHA image tags. For example, pushing branch `dev-helm-ghcr` publishes images tagged like:

```text
ghcr.io/mistervvp/logarift-frontend:dev-helm-ghcr
ghcr.io/mistervvp/logarift-frontend:sha-abc1234
```

Development Helm charts use the chart's base version with a run-number prerelease suffix, for example `0.1.0-dev.42`. The chart app version is set to the sanitized branch name, for example `dev-helm-ghcr`, so the chart defaults resolve to the matching branch image tags.

## Installing from GHCR

Authenticate first if the package visibility is private or internal:

```bash
helm registry login ghcr.io
```

Install a release chart from GHCR:

```bash
helm upgrade --install logarift oci://ghcr.io/mistervvp/charts/logarift \
  --version 1.2.3
```

Test a development branch image set by installing the development chart version from the workflow run. The chart app version defaults to the matching branch image tag:

```bash
helm upgrade --install logarift oci://ghcr.io/mistervvp/charts/logarift \
  --version 0.1.0-dev.42
```

For production-style installs, prefer immutable version or digest references over mutable branch tags. The chart supports image digests through each component's `image.digest` value.

## Required repository settings

The workflow uses the built-in `GITHUB_TOKEN` and needs these repository permissions:

- `contents: write` to upload GitHub Release assets
- `packages: write` to publish Docker images and Helm OCI charts to GHCR

If package visibility is not public, users and clusters must authenticate to GHCR before pulling images or charts.
