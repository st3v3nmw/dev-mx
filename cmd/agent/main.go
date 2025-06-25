package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/st3v3nmw/devd/internal/engine"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "check",
				Aliases: []string{"c"},
				Usage:   "check compliance to a policy",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "args",
						Usage: "key-value pairs in format key-value",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					args := cmd.Args()
					if args.Len() == 0 {
						return fmt.Errorf("policy name is required")
					}

					result := engine.CheckPolicy(args.First())
					fmt.Println(result)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
