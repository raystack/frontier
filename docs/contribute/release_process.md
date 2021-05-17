# Release Process

## Release process

For Shield maintainers, these are the concrete steps for making a new release.

1. For new major or minor release, create and check out the release branch for the new stream, e.g. `v0.6-branch`. For a patch version, check out the stream's release branch.
2. Update the [CHANGELOG.md](https://github.com/odpf/shield/blob/master/CHANGELOG.md). See the [Creating a change log](release_process.md#creating-a-change-log) guide and commit
   - Make sure to review each PR in the changelog to [flag any breaking changes and deprecation.](release_process.md#flag-breaking-changes-and-deprecations)
3. Use [npm version](https://docs.npmjs.com/cli/v6/commands/npm-version) command to generate tags, along with a version update commit.
4. Push the commits and tags. Make sure the CI passes.
   - If the CI does not pass, or if there are new patches for the release fix, repeat step 2 & 3 with release candidates until stable release is achieved.

### Creating a change log

We use an [open source change log generator](https://hub.docker.com/r/ferrarimarco/github-changelog-generator/) to generate change logs.

### Flag Breaking Changes & Deprecations

It's important to flag breaking changes and deprecation to the API for each release so that we can maintain API compatibility.

Developers should have flagged PRs with breaking changes with the `compat/breaking` label. However, it's important to double check each PR's release notes and contents for changes that will break API compatibility and manually label `compat/breaking` to PRs with undeclared breaking changes. The change log will have to be regenerated if any new labels have to be added.
