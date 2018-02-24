package commands

import (
	"github.com/tryy3/netlifyctl/commands/assets"
	"github.com/tryy3/netlifyctl/commands/deploy"
	initC "github.com/tryy3/netlifyctl/commands/init"
	"github.com/tryy3/netlifyctl/commands/login"
	"github.com/tryy3/netlifyctl/commands/middleware"
	"github.com/tryy3/netlifyctl/commands/sites"
)

func addCommands() {
	middlewares := []middleware.Middleware{
		middleware.ClientMiddleware,
		middleware.AuthMiddleware,
		middleware.LoggingMiddleware,
		middleware.DebugMiddleware,
	}

	loginMiddlewares := []middleware.Middleware{
		middleware.ClientMiddleware,
		middleware.NoAuthMiddleware,
		middleware.LoggingMiddleware,
		middleware.DebugMiddleware,
	}

	siteMiddlewares := append([]middleware.Middleware{middleware.SiteConfigMiddleware}, middlewares...)

	rootCmd.AddCommand(deploy.Setup(siteMiddlewares))
	rootCmd.AddCommand(assets.Setup(siteMiddlewares))
	rootCmd.AddCommand(sites.Setup(middlewares))
	rootCmd.AddCommand(initC.Setup(middlewares))
	rootCmd.AddCommand(login.Setup(loginMiddlewares))
	rootCmd.AddCommand(versionCmd)
}
