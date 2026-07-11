//go:build !storefront

package web

import "io/fs"

// No SPA embedded (built without -tags storefront). The storefront runs the BFF API only — used by
// `go build ./...`, tests, and local API-only runs. The release Docker build sets -tags storefront.
func distFS() (fs.FS, bool) { return nil, false }
