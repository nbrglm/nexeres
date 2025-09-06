package main

import (
	"embed"

	"github.com/nbrglm/nexeres/cmd"
)

//go:embed sqlc
var migrationsFS embed.FS

// @title NBRGLM Nexeres API Spec
// @version 0.0.1
// @description This is the NBRGLM Nexeres API Spec documentation.
// @termsOfService https://nbrglm.com/nexeres/terms

// @contact.name NBRGLM Support
// @contact.url https://nbrglm.com/support
// @contact.email contact@nbrglm.com

// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3360
// @BasePath /

// @securityDefinitions.apiKey APIKeyAuth
// @in header
// @name X-NEXERES-API-Key

// @securityDefinitions.apikey SessionHeaderAuth
// @in header
// @name X-NEXERES-Session-Token

// @securityDefinitions.apiKey RefreshHeaderAuth
// @in header
// @name X-NEXERES-Refresh-Token

// @securityDefinitions.apikey AdminHeaderAuth
// @in header
// @name X-NEXERES-Admin-Token

// @externalDocs.description NBRGLM Nexeres Documentation
// @externalDocs.url https://nbrglm.com/nexeres/docs
func main() {
	// Execute the root command.
	cmd.Exec(migrationsFS)
}
