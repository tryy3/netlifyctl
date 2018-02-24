package assets

import (
	"github.com/spf13/cobra"
	"github.com/tryy3/netlifyctl/commands/middleware"
)

func Setup(middlewares []middleware.Middleware) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "assets",
		Aliases: []string{"asset", "a"},
		Short:   "List assets attached to a site",
		Long:    "List assets attached to a site",
	}
	cmd.PersistentFlags().StringP("site-id", "s", "", "site id")

	cmd.AddCommand(setupAddCommand(middlewares))
	cmd.AddCommand(setupInfoCommand(middlewares))

	return middleware.SetupCommand(cmd, listAssets, middlewares)
}
