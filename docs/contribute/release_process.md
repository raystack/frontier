# Release Process

For Shield maintainers, these are the concrete steps for making a new release.

#### Pre-requisites

- [standard-version](https://github.com/conventional-changelog/standard-version)
- [gh cli](https://cli.github.com/manual/gh_release_create)

Run the following commands step by step to release a tag

```bash
npm run release -- --sign
git push --follow-tags origin main
gh release create <tag-version> -F CHANGELOG.md

```

**Notes:**

- For new major or minor release, create and check out the release branch for the new stream, e.g. `v0.6-branch`. For a patch version, check out the stream's release branch.
- We use an [open source change log generator](https://github.com/conventional-changelog/standard-version) to generate changelogs.
- Update the [CHANGELOG.md](https://github.com/odpf/shield/blob/master/CHANGELOG.md). See the [Creating a change log](release_process.md#creating-a-change-log) guide and commit
