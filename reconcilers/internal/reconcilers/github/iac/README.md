# Github IAC Repo Reconciler

This package replaces the github module in the old terraform setup

## Error handling

### Dapla team repo not initialized

We must manualy init the dapla team iac repoiIf reconcile fails.
This is due to we don't want to check the content of every repo to see if they have been initialized (most of them has, it is a one time thing), thus we accept the trade of of having to do this manually if a error occures which stops the reconciler from doing it.

A script is provided to ease this. Eiter via mise, or run directly with go:

- Mise: `mise run template-dapla-team-iac-repo -managed=true -team example [-output=<my path>]`
- go: Cd to reconcilers folder, and run `go run ./cmd/dapla-team-repo-init`, with same flags as for mise. Run with no flag for usage.
