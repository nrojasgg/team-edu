// Group-Compatibility subprocess plugin for ms-core.
// Compile with:
//
//	go build -o ../../plugins_dist/group_compatibility .
//
// Launched as a subprocess by the host via hashicorp/go-plugin.
package main

import (
	"ms-core/plugins_src/shared"
)

func main() {
	shared.Serve(NewGroupCompatibilityPlugin())
}
