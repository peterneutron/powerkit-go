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

3. For releases touching IOKit, SMC, Low Power Mode, or sleep behavior, capture a
   current-machine smoke readout:

   ```bash
   go run ./cmd/powerkit-cli all
   ```

   Note any OS-specific capability changes in the release notes.
4. Prepare concise release notes from the merged commits.
5. Tag the release on `master`:

   ```bash
   git tag vX.Y.Z
   ```

6. Push `master` and the new tag.
7. Publish release notes on the hosting platform for that tag.
8. If downstream repos pin `powerkit-go`, bump them after the tag exists.

## Versioning Notes

- use semver tags on `master`
- additive API changes may ship as patch releases
- breaking exported API or JSON contract changes require a minor or major bump
- JSON `schema_version` and Go module semver are separate version tracks

## CI Expectations

CI should validate `master` and pull requests targeting `master`. If the branch model changes again, update the workflows and this document together.
