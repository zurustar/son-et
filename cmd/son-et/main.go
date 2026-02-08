package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/zurustar/son-et/pkg/app"
)

//go:embed titles soundfonts
var embeddedTitles embed.FS

func main() {
	application := app.New(embeddedTitles)
	if err := application.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
