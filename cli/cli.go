package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/runabol/tork"
	"github.com/runabol/tork/bootstrap"
	"github.com/runabol/tork/conf"
	ucli "github.com/urfave/cli/v2"
)

func Run() error {
	app := &ucli.App{
		Name:     "tork",
		Usage:    "a distributed workflow engine",
		Before:   before,
		Commands: commands(),
	}
	return app.Run(os.Args)
}

func before(ctx *ucli.Context) error {
	displayBanner()
	return nil
}

func commands() []*ucli.Command {
	return []*ucli.Command{
		runCmd(),
		migrationCmd(),
		healthCmd(),
	}
}

func runCmd() *ucli.Command {
	return &ucli.Command{
		Name:      "run",
		Usage:     "Run Tork",
		UsageText: "tork run mode (standalone|coordinator|worker)",
		Action: func(ctx *ucli.Context) error {
			mode := ctx.Args().First()
			if mode == "" {
				if err := ucli.ShowSubcommandHelp(ctx); err != nil {
					return err
				}
				fmt.Println("missing required argument: mode")
				os.Exit(1)

			}
			return bootstrap.Start(bootstrap.Mode(ctx.Args().First()))
		},
	}
}

func migrationCmd() *ucli.Command {
	return &ucli.Command{
		Name:  "migration",
		Usage: "Run the db migration script",
		Action: func(ctx *ucli.Context) error {
			return bootstrap.Start(bootstrap.ModeMigration)
		},
	}
}

func healthCmd() *ucli.Command {
	return &ucli.Command{
		Name:   "health",
		Usage:  "Perform a health check",
		Action: health,
	}
}

func displayBanner() {
	mode := conf.StringDefault("cli.banner.mode", "console")
	if mode == "off" {
		return
	}
	banner := color.WhiteString(fmt.Sprintf(`
 _______  _______  ______    ___   _ 
|       ||       ||    _ |  |   | | |
|_     _||   _   ||   | ||  |   |_| |
  |   |  |  | |  ||   |_||_ |      _|
  |   |  |  |_|  ||    __  ||     |_ 
  |   |  |       ||   |  | ||    _  |
  |___|  |_______||___|  |_||___| |_|

 %s (%s)
`, tork.Version, tork.GitCommit))

	if mode == "console" {
		fmt.Println(banner)
	} else {
		log.Info().Msg(banner)
	}
}

func health(_ *ucli.Context) error {
	chk, err := http.Get(fmt.Sprintf("%s/health", conf.StringDefault("endpoint", "http://localhost:8000")))
	if err != nil {
		return err
	}
	if chk.StatusCode != http.StatusOK {
		return errors.Errorf("Health check failed. Status Code: %d", chk.StatusCode)
	}
	body, err := io.ReadAll(chk.Body)
	if err != nil {
		return errors.Wrapf(err, "error reading body")
	}

	type resp struct {
		Status string `json:"status"`
	}
	r := resp{}

	if err := json.Unmarshal(body, &r); err != nil {
		return errors.Wrapf(err, "error unmarshalling body")
	}

	fmt.Printf("Status: %s\n", r.Status)

	return nil
}
