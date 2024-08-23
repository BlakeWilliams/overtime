package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/blakewilliams/overtime/generator"
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "package name",
						Usage: "The name of the package to generate",
					},
					&cli.StringFlag{
						Name:  "directory",
						Usage: "The directory to create the package in. Pro Tip™ use '..' + `go generate` in your impl.go file",
					},
				},
				Usage: "Generate a REST gateway from a schema",
				Action: func(c *cli.Context) error {
					if c.Args().Len() < 1 {
						return fmt.Errorf("You must pass a schema file to generate")
					}

					log.Println("Generating a REST gateway from the provided schema...")
					schemaFile := c.Args().First()
					fmt.Println(os.Getwd())
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
					gen := generator.NewGo(graph)
					gen.PackageName = "overtime"
					if packageName := c.String("package name"); packageName != "" {
						gen.PackageName = packageName
					}

					rootPath := gen.PackageName
					if directory := c.String("directory"); directory != "" {
						rootPath = path.Join(directory, gen.PackageName)
					}

					_ = os.Mkdir(gen.PackageName, 0755)
					if err := writeFile(path.Join(rootPath, "types.go"), gen.GenerateTypes()); err != nil {
						return err
					}
					if err := writeFile(path.Join(rootPath, "resolvers.go"), gen.Resolvers()); err != nil {
						return err
					}
					if err := writeFile(path.Join(rootPath, "controllers.go"), gen.GenerateControllers()); err != nil {
						return err
					}
					if err := writeFile(path.Join(rootPath, "coordinator.go"), gen.Coordinator()); err != nil {
						return err
					}

					if _, err := os.Stat(path.Join(rootPath, "impl.go")); os.IsNotExist(err) {
						if err := writeFile(path.Join(rootPath, "impl.go"), gen.Root()); err != nil {
							return err
						}
					}

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

func writeFile(path string, r io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("Failed to write to file %s: %w", path, err)
	}

	// run go fmt on the file
	cmd := exec.Command("gofmt", "-w", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to run gofmt on file %s: %w", path, err)
	}

	fmt.Printf("Created %s\n", path)

	return nil
}
