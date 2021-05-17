# Development Guide

This guide is targeted at developers looking to contribute to Shield. Check these sections before you start contributing:

- [Concepts](development_guide.md#concepts)
- [Making a Pull Request](development_guide.md#making-a-pull-request)

## Concepts

Before contributing, read through the entire [concepts](https://odpf.gitbook.io/shield/concepts/overview) section to understand Shield better.

## Making a Pull Request

#### Incorporating upstream changes from main

Our preference is the use of `git rebase` instead of `git merge` : `git pull -r`

#### Signing commits

Commits have to be signed before they are allowed to be merged into the codebase:

```bash
# Include -s flag to signoff
git commit -s -m "My first commit"
```

#### Good practices to keep in mind

- Follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary) while composing our commit messages.
- Add `WIP:` to PR name if more work needs to be done prior to review
- Avoid `force-pushing` as it makes reviewing difficult

**Managing CI-test failures**

- GitHub runner tests
  - Click `checks` tab to analyse failed tests
