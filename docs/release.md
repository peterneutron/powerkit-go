# Release Process

This repo uses `master` as the trunk and tagged release branch.

## Branch Model

- land feature and fix commits on `master`
- keep `master` releasable and tag only from `master`
- avoid parallel long-lived release branches unless they provide real day-to-day value

## Patch Release Checklist

1. Finalize the release candidate on `master`.
2. Run verification from repo root:

   ```bash
   make verify
   ```

3. Update [CHANGELOG.md](../CHANGELOG.md) under `Unreleased`.
4. Tag the release on `master`:

   ```bash
   git tag vX.Y.Z
   ```

5. Push `master` and the new tag.
6. If downstream repos pin `powerkit-go`, bump them after the tag exists.

## Versioning Notes

- use semver tags on `master`
- additive API changes may ship as patch releases
- breaking exported API or JSON contract changes require a minor or major bump
- JSON `schema_version` and Go module semver are separate version tracks

## CI Expectations

CI should validate `master` and pull requests targeting `master`. If the branch model changes again, update the workflows and this document together.
