/*
Package main implements staticlint.

Staticlint is a multichecker that combines:

  - standard go/analysis passes
  - all SA analyzers from Staticcheck
  - stylecheck analyzers
  - quickfix analyzers
  - errcheck
  - ineffassign
  - custom osexit analyzer

Usage:

	go build -o staticlint ./cmd/staticlint

	staticlint ./...

Analyzer osexit reports direct calls to os.Exit
inside function main of package main.
*/
package main
