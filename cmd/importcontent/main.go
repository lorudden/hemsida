package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/lorudden/hemsida/internal/integrations/wordpress"
)

func main() {
	var (
		contentDir = flag.String("content-dir", "data/content", "directory containing content manifests")
		mediaDir   = flag.String("media-dir", "data/media", "directory containing downloaded legacy media assets")
		dryRun     = flag.Bool("dry-run", false, "fetch and parse content without writing manifest files")
		timeout    = flag.Duration("timeout", 20*time.Second, "HTTP timeout for WordPress requests")
	)
	flag.Parse()

	importer := wordpress.NewImporterWithOptions(wordpress.DefaultHTTPClient(*timeout), wordpress.ImporterOptions{
		MediaRoot: *mediaDir,
	})
	ctx := context.Background()

	files := wordpress.DefaultCollectionFiles(*contentDir)
	for _, path := range files {
		result, err := importer.ImportCollectionFile(ctx, path, *dryRun)
		if err != nil {
			log.Fatalf("import failed for %s: %v", path, err)
		}

		mode := "updated"
		if *dryRun {
			mode = "checked"
		}
		fmt.Printf("%s %s (%d entries)\n", mode, result.Path, result.EntriesUpdated)
	}
}
