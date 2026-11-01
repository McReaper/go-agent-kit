package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/McReaper/go-agent-kit/internal/platform/ghapp"
	"github.com/spf13/cobra"
)

var errNoApp = errors.New("give the app's Client ID and key: --client-id and --key, or KIT_APP_CLIENT_ID and KIT_APP_KEY")

type appOptions struct {
	clientID     string
	key          string
	installation int64
}

func appCommand(out io.Writer) *cobra.Command {
	o := &appOptions{clientID: os.Getenv("KIT_APP_CLIENT_ID"), key: os.Getenv("KIT_APP_KEY")}
	if id, err := strconv.ParseInt(os.Getenv("KIT_APP_INSTALLATION"), 10, 64); err == nil {
		o.installation = id
	}
	app := &cobra.Command{Use: "app", Short: "Act as the agent's GitHub App"}
	app.PersistentFlags().StringVar(&o.clientID, "client-id", o.clientID, "the app's Client ID (KIT_APP_CLIENT_ID)")
	app.PersistentFlags().StringVar(&o.key, "key", o.key, "the app's private key file (KIT_APP_KEY)")
	app.AddCommand(installationsCommand(o, out), execCommand(o))
	return app
}

func (o *appOptions) jwt() (string, error) {
	if o.clientID == "" || o.key == "" {
		return "", errNoApp
	}
	key, err := ghapp.LoadKey(o.key)
	if err != nil {
		return "", err
	}
	return ghapp.JWT(o.clientID, key, time.Now())
}

func installationsCommand(o *appOptions, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "installations",
		Short: "List where the app is installed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			jwt, err := o.jwt()
			if err != nil {
				return err
			}
			list, err := ghapp.NewClient().Installations(cmd.Context(), jwt)
			if err != nil {
				return err
			}
			for _, i := range list {
				if _, err := fmt.Fprintf(out, "%d\t%s\t%s\n", i.ID, i.Account.Login, i.RepositorySelection); err != nil {
					return fmt.Errorf("print installations: %w", err)
				}
			}
			return nil
		},
	}
}

func execCommand(o *appOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec -- command [args...]",
		Short: "Run a command as the app: gh and git push with its token, commits by its bot",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, argv []string) error {
			return runAsApp(cmd.Context(), o, argv)
		},
	}
	cmd.Flags().Int64Var(&o.installation, "installation", o.installation, "the installation's id (KIT_APP_INSTALLATION)")
	return cmd
}

func runAsApp(ctx context.Context, o *appOptions, argv []string) error {
	jwt, err := o.jwt()
	if err != nil {
		return err
	}
	c := ghapp.NewClient()
	tok, err := c.Token(ctx, jwt, o.installation)
	if err != nil {
		return err
	}
	slug, err := c.Slug(ctx, jwt)
	if err != nil {
		return err
	}
	id := ghapp.Identity{Login: slug + "[bot]"}
	if id.ID, err = c.UserID(ctx, id.Login); err != nil {
		return err
	}
	return ghapp.Exec(ctx, argv, tok, id)
}
