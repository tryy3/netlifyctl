package deploy

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	netlify "github.com/netlify/open-api/go/porcelain"
	"github.com/tryy3/netlifyctl/commands/middleware"
	"github.com/tryy3/netlifyctl/configuration"
	"github.com/tryy3/netlifyctl/context"
	"github.com/tryy3/netlifyctl/operations"
	"github.com/tryy3/netlifyctl/ui"
)

type deployCmd struct {
	base      string
	title     string
	functions string
	siteID    string
	siteName  string
	draft     bool
	browser   bool
}

func Setup(middlewares []middleware.Middleware) *cobra.Command {
	cmd := &deployCmd{}
	ccmd := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"deploys", "d"},
		Short:   "Deploy your site",
		Long:    "Deploy your site",
	}
	ccmd.Flags().StringVarP(&cmd.base, "base-directory", "b", "", "directory to publish")
	ccmd.Flags().StringVarP(&cmd.title, "message", "m", "", "message for the deploy title")
	ccmd.Flags().BoolVarP(&cmd.draft, "draft", "d", false, "draft deploy, not published in production")
	ccmd.Flags().StringVarP(&cmd.functions, "functions", "f", "", "function directory to deploy")
	ccmd.Flags().StringVarP(&cmd.siteID, "site-id", "s", "", "explicitly set a site id instead of relying on configuration")
	ccmd.Flags().StringVarP(&cmd.siteName, "name", "n", "", "search a site by its name instead of relying on configuration")
	ccmd.Flags().BoolVarP(&cmd.browser, "browser", "o", false, "open the url in the browser after deploying")

	return middleware.SetupCommand(ccmd, cmd.deploySite, middlewares)
}

func (dc *deployCmd) deploySite(ctx context.Context, cmd *cobra.Command, args []string) error {
	conf := context.GetSiteConfig(ctx)
	if conf.Settings.ID == "" {
		return errors.New("Failed to load site configuration")
	}

	draft, err := cmd.Flags().GetBool("draft")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get string flag: 'draft'")
	}

	fs, err := cmd.Flags().GetString("functions")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get string flag: 'functions'")
	}
	if fs == "" && conf.Build.Functions != "" {
		fs = conf.Build.Functions
	}

	dir := baseDeploy(cmd, conf)

	obs := operations.NewDeployObserver()

	client := context.GetClient(ctx)
	options := netlify.DeployOptions{
		Observer:     obs,
		SiteID:       conf.Settings.ID,
		Dir:          dir,
		IsDraft:      draft,
		FunctionsDir: fs,
		Title:        dc.title,
	}

	logrus.WithFields(logrus.Fields{
		"site":  options.SiteID,
		"path":  options.Dir,
		"draft": options.IsDraft}).Debug("deploying site")

	d, err := client.DeploySite(ctx, options)
	if err != nil {
		return err
	}

	if len(d.Required) > 0 {
		ready, err := client.WaitUntilDeployReady(ctx, d)
		if err != nil {
			return err
		}
		d = ready
	}

	obs.Finish()

	u := d.SslURL
	if d.Context != "production" {
		u = d.DeploySslURL
	}
	fmt.Printf("Deploy done  %s\n", ui.WorldCheck())
	ui.Bold("    %s\n", u)

	if browser, err := cmd.Flags().GetBool("browser"); err == nil && browser {
		open(u)
	}

	return nil
}

func baseDeploy(cmd *cobra.Command, conf *configuration.Configuration) string {
	bd, err := cmd.Flags().GetString("base-directory")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get string flag: 'base-directory'")
	}

	if bd != "" {
		return bd
	}

	s := conf.Settings
	path := s.Path

	if path == "" && conf.Build.Publish != "" {
		path = conf.Build.Publish
	}

	if path == "" {
		path, err = ui.AskForInput("What path would you like deployed?", ".")
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get deploy path")
		}

		logrus.Debugf("Got new path from the user %s", s.Path)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(conf.Root(), path)
		logrus.Debugf("Relative path detected, going to deploy: '%s'", path)
	}

	s.Path = path

	return path
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
