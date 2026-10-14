package main

import (
	"ms-core/plugins_src/shared"
)

func main() {
	shared.Serve(NewMyPlugin())
}
