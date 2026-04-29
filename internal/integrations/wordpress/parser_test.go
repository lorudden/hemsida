package wordpress

import (
	"strings"
	"testing"
)

func TestParsePageExtractsSummaryAndImages(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Lörudden - Lörans Hamnförening</title>
				<meta name="description" content="Beskrivning från metadata">
				<meta property="article:published_time" content="2024-06-20T10:30:00Z">
			</head>
			<body>
				<div class="site-header">
					<p>Kort puff som inte tillhör innehållet.</p>
				</div>
				<main class="entry-content">
					<h1>Lörudden</h1>
					<p>Första stycket om Lörudden med lite mer sammanhang för att bli en bra ingress i den nya sajten.</p>
					<div class="hero slider">
						<img src="/wp-content/uploads/hero.jpg" alt="Hamnen i kvällsljus"/>
						<img src="https://example.com/absolute.jpg" alt="Absolut bild"/>
					</div>
					<figure>
						<img src="/wp-content/uploads/detail.jpg" alt="Detaljbild"/>
						<figcaption>Detaljbild över hamnen</figcaption>
					</figure>
				</main>
			</body>
		</html>`

	page, err := ParsePage(strings.NewReader(html), "https://example.se/?page_id=69")
	if err != nil {
		t.Fatalf("ParsePage() error = %v", err)
	}

	if got := page.Title; got != "Lörudden" {
		t.Fatalf("page.Title = %q, want Lörudden", got)
	}
	if got := page.Summary; got != "Första stycket om Lörudden med lite mer sammanhang för att bli en bra ingress i den nya sajten." {
		t.Fatalf("page.Summary = %q, want main content paragraph", got)
	}
	if page.PublishedAt == nil {
		t.Fatalf("page.PublishedAt = nil, want parsed time")
	}
	if got := len(page.ImageCollections); got != 2 {
		t.Fatalf("len(page.ImageCollections) = %d, want 2", got)
	}
	if got := page.ImageCollections[0].Kind; got != "hero" {
		t.Fatalf("page.ImageCollections[0].Kind = %q, want hero", got)
	}
	if got := len(page.ImageCollections[0].Items); got != 2 {
		t.Fatalf("len(page.ImageCollections[0].Items) = %d, want 2", got)
	}
	if got := page.ImageCollections[1].Items[0].Caption; got != "Detaljbild över hamnen" {
		t.Fatalf("page.ImageCollections[1].Items[0].Caption = %q", got)
	}
}

func TestParsePageFallsBackToParagraph(t *testing.T) {
	html := `
		<html>
			<head><title>Nyheter</title></head>
			<body><p>Detta ar ingressen.</p></body>
		</html>`

	page, err := ParsePage(strings.NewReader(html), "https://example.se/?page_id=80")
	if err != nil {
		t.Fatalf("ParsePage() error = %v", err)
	}

	if got := page.Summary; got != "Detta ar ingressen." {
		t.Fatalf("page.Summary = %q, want paragraph fallback", got)
	}
}

func TestParsePagePrefersContentRootOverSidebar(t *testing.T) {
	html := `
		<html>
			<head><title>Nyheter</title></head>
			<body>
				<aside><p>Sidokolumn med irrelevant text.</p></aside>
				<div class="entry-content">
					<p>Detta ar den riktiga ingresstexten for innehallet.</p>
				</div>
			</body>
		</html>`

	page, err := ParsePage(strings.NewReader(html), "https://example.se/news")
	if err != nil {
		t.Fatalf("ParsePage() error = %v", err)
	}

	if got := page.Summary; got != "Detta ar den riktiga ingresstexten for innehallet." {
		t.Fatalf("page.Summary = %q", got)
	}
}

func TestParsePageFiltersNonContentImages(t *testing.T) {
	html := `
		<html>
			<body>
				<div class="entry-content">
					<p>Detta ar innehallet med en riktig bild.</p>
					<img src="https://secure.gravatar.com/avatar/123?s=40" alt="avatar"/>
					<img src="/wp-content/uploads/photo.jpg" alt="foto"/>
				</div>
			</body>
		</html>`

	page, err := ParsePage(strings.NewReader(html), "https://example.se/page")
	if err != nil {
		t.Fatalf("ParsePage() error = %v", err)
	}

	if got := len(page.ImageCollections); got != 1 {
		t.Fatalf("len(page.ImageCollections) = %d, want 1", got)
	}
	if got := len(page.ImageCollections[0].Items); got != 1 {
		t.Fatalf("len(page.ImageCollections[0].Items) = %d, want 1", got)
	}
}

func TestParseNewsArchiveExtractsPosts(t *testing.T) {
	html := `
		<html>
			<body>
				<div id="post-2113" class="post-2113 post type-post status-publish">
					<h2 class="entry-title"><a href="https://example.se/?p=2113">Brandövning</a></h2>
					<div class="entry-meta">
						<span class="entry-date">juli 6, 2025</span>
					</div>
					<div class="entry-content">
						<p>Detta ar en nyhetssammanfattning med tillrackligt mycket text for att valjas.</p>
						<figure class="wp-block-image size-large">
							<a href="http://example.se/wp-content/uploads/2025/07/Brandovning-2025_1.jpg">
								<img src="http://example.se/wp-content/uploads/2025/07/Brandovning-2025_1-724x1024.jpg" alt="Brand"/>
							</a>
						</figure>
					</div>
				</div>
			</body>
		</html>`

	entries, err := ParseNewsArchive(strings.NewReader(html), "https://example.se/?page_id=80")
	if err != nil {
		t.Fatalf("ParseNewsArchive() error = %v", err)
	}

	if got := len(entries); got != 1 {
		t.Fatalf("len(entries) = %d, want 1", got)
	}
	if got := entries[0].Slug; got != "post-2113" {
		t.Fatalf("entries[0].Slug = %q, want post-2113", got)
	}
	if entries[0].PublishedAt == nil {
		t.Fatalf("entries[0].PublishedAt = nil")
	}
	if got := entries[0].ImageCollections[0].Items[0].SourceURL; got != "https://example.se/wp-content/uploads/2025/07/Brandovning-2025_1.jpg" {
		t.Fatalf("entries[0].ImageCollections[0].Items[0].SourceURL = %q", got)
	}
}
