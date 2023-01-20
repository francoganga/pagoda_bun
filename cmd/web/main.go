package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/francoganga/pagoda_bun/cmd/web/migrations"
	"github.com/francoganga/pagoda_bun/pkg/routes"
	"github.com/francoganga/pagoda_bun/pkg/services"
	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name: "Tramites internos",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "env",
				Value: "dev",
				Usage: "environment",
			},
		},
		Commands: []*cli.Command{
			startCommand(),
			migrateCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func startCommand() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "start application",
		Action: func(*cli.Context) error {

			c := services.NewContainer()

			defer func() {
				if err := c.Shutdown(); err != nil {
					c.Web.Logger.Fatal(err)
				}
			}()

			// Build the router
			routes.BuildRouter(c)

			// Start the server
			go func() {
				srv := http.Server{
					Addr:         fmt.Sprintf("%s:%d", c.Config.HTTP.Hostname, c.Config.HTTP.Port),
					Handler:      c.Web,
					ReadTimeout:  c.Config.HTTP.ReadTimeout,
					WriteTimeout: c.Config.HTTP.WriteTimeout,
					IdleTimeout:  c.Config.HTTP.IdleTimeout,
				}

				if c.Config.HTTP.TLS.Enabled {
					certs, err := tls.LoadX509KeyPair(c.Config.HTTP.TLS.Certificate, c.Config.HTTP.TLS.Key)
					if err != nil {
						c.Web.Logger.Fatalf("cannot load TLS certificate: %v", err)
					}

					srv.TLSConfig = &tls.Config{
						Certificates: []tls.Certificate{certs},
					}
				}

				if err := c.Web.StartServer(&srv); err != http.ErrServerClosed {
					c.Web.Logger.Fatalf("shutting down the server: %v", err)
				}
			}()

			// Start the scheduler service to queue periodic tasks
			go func() {
				if err := c.Tasks.StartScheduler(); err != nil {
					c.Web.Logger.Fatalf("scheduler shutdown: %v", err)
				}
			}()

			// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			signal.Notify(quit, os.Kill)
			<-quit
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := c.Web.Shutdown(ctx); err != nil {
				c.Web.Logger.Fatal(err)
			}

			return nil
		},
	}
}

var migrateCommand = &cli.Command{
	Name:  "db",
	Usage: "manage database",
	Subcommands: []*cli.Command{
		{
			Name:  "migrate",
			Usage: "migrate database",
			Action: func(c *cli.Context) error {

				container := services.NewContainer()

				defer container.Shutdown()

				migrator := migrate.NewMigrator(container.Bun, migrations.Migrations)

				err := migrator.Init(context.Background())

				if err != nil {
					return err
				}

				group, err := migrator.Migrate(context.Background())
				if err != nil {
					return err
				}

				if group.ID == 0 {
					fmt.Printf("there are no new migrations to run\n")
					return nil
				}

				fmt.Printf("migrated to %s\n", group)
				return nil

			},
		},
		{
			Name:  "create_sql",
			Usage: "create up and down SQL migrations",
			Action: func(c *cli.Context) error {

				container := services.NewContainer()

				defer container.Shutdown()

				migrator := migrate.NewMigrator(container.Bun, migrations.Migrations)

				name := strings.Join(c.Args().Slice(), "_")
				files, err := migrator.CreateSQLMigrations(context.Background(), name)
				if err != nil {
					return err
				}

				for _, mf := range files {
					fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
				}

				return nil
			},
		},
	},
}
