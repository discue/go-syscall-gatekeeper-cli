// The CLI entrypoint previously lived in this repository (see `main.go`).
//
// The CLI has been moved to a separate project. That project now imports
// this library and calls into the public API exposed in the `gatekeeper`
// package (see `gatekeeper.Start` / `gatekeeper.StartAndWait`).

package main
