package main

import (
	"fmt"
	"log"
	"os"

	"github.com/blakewilliams/overtime/internal/parser"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "overtime",
		Usage: "A tool for generating a (federated) REST gateway from a schema",

		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Generates a basic config and schema for hosting a gateway",
				Action: func(c *cli.Context) error {
					fmt.Println("Initializing a new overtime project")
					return nil
				},
			},
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "Generate a REST gateway from a schema",
				Action: func(c *cli.Context) error {
					if c.Args().Len() < 1 {
						return fmt.Errorf("You must pass a schema file to generate")
					}

					log.Println("Generating a REST gateway from the provided schema...")
					schemaFile := c.Args().First()
					if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
						return fmt.Errorf("The schema file %s does not exist", schemaFile)
					}

					// TODO use io.Reader instead of os.ReadFile
					contents, err := os.ReadFile(schemaFile)
					if err != nil {
						return fmt.Errorf("Failed to read the schema file %s: %w", schemaFile, err)
					}

					graph, err := parser.Parse(string(contents))
					if err != nil {
						return fmt.Errorf("Failed to parse the schema: %w", err)
					}

					fmt.Println(graph)

					// check if file exists first
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		defer func() {
			os.Exit(1)
		}()
	}
}
