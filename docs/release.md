# Release Process

This repo uses `dev` as the integration branch and `master` as the tagged release branch.

## Branch Model

- land feature and fix commits on `dev`
- keep `master` releasable and tag only from `master`
- keep `dev` and `master` synchronized after releases so they do not drift by tree content

## Patch Release Checklist

1. Finalize the release candidate on `dev`.
2. Run verification from repo root:

   ```bash
   make verify
   ```

3. Update [CHANGELOG.md](../CHANGELOG.md) under `Unreleased`.
4. Merge `dev` into `master` without adding extra release-only code changes.
5. Tag the release on `master`:

   ```bash
   git tag vX.Y.Z
   ```

6. Push `master`, `dev`, and the new tag.
7. If downstream repos pin `powerkit-go`, bump them after the tag exists.

## Versioning Notes

- use semver tags on `master`
- additive API changes may ship as patch releases
- breaking exported API or JSON contract changes require a minor or major bump
- JSON `schema_version` and Go module semver are separate version tracks

## CI Expectations

CI should validate both `dev` and `master`. If the documented branch model changes, update the workflows and this document together.
