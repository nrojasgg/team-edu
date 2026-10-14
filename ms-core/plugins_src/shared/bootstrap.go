package shared

import (
	"os"

	"ms-core/internal/core"
	"ms-core/internal/database"
)

func Serve(plugin core.Plugin) {
	if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
		database.ConnectDB(mongoURI)
		defer database.DisconnectDB()
	}

	core.ServePlugin(plugin)
}
