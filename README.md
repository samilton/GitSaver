# GitSaver

## Thoughts

### How do we handle concurrency?
This should run as a Lambda or KNative function and be stateless.
However, it needs to handle multiple pushes to `main` in quick succession.
**Question**: Should we place new webhook payloads on a Kafka topic and then process them with a series of runners?
Maybe we can use [valkey](https://github.com/valkey-io/valkey) as a queue instead and work on them sequentially...

### Storage
We should store the backed up repositories in a single bucket but organized by organization `infra/gitsaver`
We should enable versioning enabled on the bucket

### Status
We should provide a status page that that can be displayed that shows the repo, the last update, with the sha and commit message

### Failures
- We should retry failures but alert when we've failed 3 times
- Failures should create an issue in GitHub issues for the GitSaver repo

## Random Thoughts

### How do we handle Repositories with Submodules?
I think we can ignore the fact that they have submodules.
This may lead to a point where a repo that depends on a submodule that has never been backed up would be unable to build

## Decision Log
1. We will start by having a goroutine that can be aborted before adding too much infra
