package wordpress

import (
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/lorudden/hemsida/internal/content/catalog"
	"golang.org/x/net/html"
)

// ParsedPage is the importer-oriented representation of a WordPress page.
type ParsedPage struct {
	Title            string
	Summary          string
	PublishedAt      *time.Time
	ImageCollections []catalog.ImageCollection
}

// ParsedArchiveEntry represents a single post teaser or article discovered on an archive page.
type ParsedArchiveEntry struct {
	Title            string
	Slug             string
	SourcePageURL    string
	Summary          string
	PublishedAt      *time.Time
	ImageCollections []catalog.ImageCollection
}

// ParsePage extracts visible metadata and externally linked image references from HTML.
func ParsePage(r io.Reader, baseURL string) (ParsedPage, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return ParsedPage{}, err
	}

	contentRoot := findContentRoot(doc)
	title := normalizeWhitespace(firstNonEmpty(
		findHeading(contentRoot, "h1"),
		findMetaProperty(doc, "og:title"),
		findTitle(doc),
	))
	summary := normalizeWhitespace(firstNonEmpty(
		firstMeaningfulParagraph(contentRoot),
		findMetaDescription(doc),
		firstParagraph(contentRoot),
		firstParagraph(doc),
	))

	return ParsedPage{
		Title:            title,
		Summary:          summary,
		PublishedAt:      findPublishedAt(doc, contentRoot),
		ImageCollections: findImageCollections(contentRoot, baseURL),
	}, nil
}

// ParseNewsArchive extracts individual news entries from a WordPress archive-like page.
func ParseNewsArchive(r io.Reader, baseURL string) ([]ParsedArchiveEntry, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	postNodes := findArchivePosts(doc)
	entries := make([]ParsedArchiveEntry, 0, len(postNodes))
	for _, postNode := range postNodes {
		entry := parseArchiveEntry(postNode, baseURL)
		if entry.Title == "" || entry.SourcePageURL == "" {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func findContentRoot(root *html.Node) *html.Node {
	var best *html.Node
	bestScore := -1

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			score := contentNodeScore(node)
			if score > bestScore {
				best = node
				bestScore = score
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)

	if best == nil {
		return root
	}
	return best
}

func contentNodeScore(node *html.Node) int {
	if node.Type != html.ElementNode {
		return 0
	}

	score := 0
	switch node.Data {
	case "main":
		score += 8
	case "article":
		score += 7
	case "section":
		score += 2
	case "div":
		score += 1
	}

	classes := " " + nodeClass(node) + " "
	id := " " + attr(node, "id") + " "

	for _, hint := range []string{
		" entry-content ",
		" post-content ",
		" page-content ",
		" site-content ",
		" main-content ",
		" content-area ",
		" content ",
		" post ",
		" article ",
	} {
		if strings.Contains(classes, hint) {
			score += 10
		}
		if strings.Contains(id, hint) {
			score += 10
		}
	}

	for _, exactHint := range []string{"entry-content", "post-content", "page-content"} {
		if hasClass(node, exactHint) || strings.EqualFold(attr(node, "id"), exactHint) {
			score += 25
		}
	}

	for _, penalty := range []string{"comment", "comments", "widget", "sidebar", "footer", "author-bio"} {
		if hasClass(node, penalty) || strings.Contains(id, " "+penalty+" ") {
			score -= 20
		}
	}

	score += min(descendantCount(node, "p"), 6)
	score += min(descendantCount(node, "img"), 4)
	return score
}

func findTitle(root *html.Node) string {
	var title string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if title != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == "title" {
			title = normalizeWhitespace(textContent(node))
			title = strings.TrimSuffix(title, " - Lörans Hamnförening")
			title = strings.TrimSpace(title)
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return title
}

func findMetaDescription(root *html.Node) string {
	description := findMetaByName(root, "description")
	if description != "" {
		return description
	}
	return findMetaProperty(root, "og:description")
}

func findMetaByName(root *html.Node, name string) string {
	var value string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if value != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == "meta" && strings.EqualFold(attr(node, "name"), name) {
			value = attr(node, "content")
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return value
}

func findMetaProperty(root *html.Node, property string) string {
	var value string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if value != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == "meta" && strings.EqualFold(attr(node, "property"), property) {
			value = attr(node, "content")
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return value
}

func firstMeaningfulParagraph(root *html.Node) string {
	text := firstParagraph(root)
	if len([]rune(text)) >= 40 {
		return text
	}
	return ""
}

func firstParagraph(root *html.Node) string {
	var paragraph string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if paragraph != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == "p" {
			text := normalizeWhitespace(textContent(node))
			if text != "" {
				paragraph = text
				return
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return paragraph
}

func findHeading(root *html.Node, tag string) string {
	var heading string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if heading != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == tag {
			heading = normalizeWhitespace(textContent(node))
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return heading
}

func findPublishedAt(doc *html.Node, contentRoot *html.Node) *time.Time {
	for _, candidate := range []string{
		findMetaProperty(doc, "article:published_time"),
		findMetaByName(doc, "article:published_time"),
		findTextByClass(contentRoot, "entry-date"),
		findTimeDatetime(contentRoot),
		findTimeDatetime(doc),
	} {
		if parsed := parseTime(candidate); parsed != nil {
			return parsed
		}
	}
	return nil
}

func findTextByClass(root *html.Node, className string) string {
	node := findDescendantByClass(root, className)
	if node == nil {
		return ""
	}
	return normalizeWhitespace(textContent(node))
}

func findTimeDatetime(root *html.Node) string {
	var value string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if value != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == "time" {
			value = attr(node, "datetime")
			if value == "" {
				value = normalizeWhitespace(textContent(node))
			}
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return value
}

func parseTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	value = translateSwedishMonth(value)
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"January 2, 2006",
		"2 January 2006",
		"02 January 2006",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return &t
		}
	}
	return nil
}

func translateSwedishMonth(value string) string {
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return value
	}

	months := map[string]string{
		"januari":   "January",
		"februari":  "February",
		"mars":      "March",
		"april":     "April",
		"maj":       "May",
		"juni":      "June",
		"juli":      "July",
		"augusti":   "August",
		"september": "September",
		"oktober":   "October",
		"november":  "November",
		"december":  "December",
	}

	first := strings.Trim(fields[0], ",")
	if translated, ok := months[strings.ToLower(first)]; ok {
		fields[0] = translated
	}

	return strings.Join(fields, " ")
}

func findImageCollections(root *html.Node, baseURL string) []catalog.ImageCollection {
	groupNodes := findCandidateImageGroups(root)
	seen := make(map[string]struct{})
	collections := make([]catalog.ImageCollection, 0, len(groupNodes)+1)

	for _, group := range groupNodes {
		images := imagesFromNode(group, baseURL, seen)
		if len(images) == 0 {
			continue
		}
		collections = append(collections, catalog.ImageCollection{
			Title:      imageCollectionTitle(group),
			Kind:       imageCollectionKind(group, len(images)),
			ImportMode: catalog.ImageImportModeExternal,
			Items:      images,
		})
	}

	remaining := imagesFromNode(root, baseURL, seen)
	if len(remaining) > 0 {
		collections = append(collections, catalog.ImageCollection{
			Title:      fallbackImageCollectionTitle(len(collections), len(remaining)),
			Kind:       fallbackImageCollectionKind(len(remaining)),
			ImportMode: catalog.ImageImportModeExternal,
			Items:      remaining,
		})
	}

	return collections
}

func findArchivePosts(root *html.Node) []*html.Node {
	posts := make([]*html.Node, 0)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if isArchivePostNode(node) {
			posts = append(posts, node)
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return posts
}

func isArchivePostNode(node *html.Node) bool {
	if node == nil || node.Type != html.ElementNode {
		return false
	}
	return hasClass(node, "post") && hasClass(node, "type-post")
}

func parseArchiveEntry(postNode *html.Node, baseURL string) ParsedArchiveEntry {
	contentRoot := findDescendantByClass(postNode, "entry-content")
	if contentRoot == nil {
		contentRoot = postNode
	}

	titleNode := findDescendantByClass(postNode, "entry-title")
	title := normalizeWhitespace(textContent(titleNode))
	sourceURL := ""
	if titleNode != nil {
		if anchor := findFirstElement(titleNode, "a"); anchor != nil {
			sourceURL = resolveURL(baseURL, attr(anchor, "href"))
			if title == "" {
				title = normalizeWhitespace(textContent(anchor))
			}
		}
	}

	return ParsedArchiveEntry{
		Title:            title,
		Slug:             slugFromURL(sourceURL),
		SourcePageURL:    sourceURL,
		Summary:          normalizeWhitespace(firstNonEmpty(firstMeaningfulParagraph(contentRoot), firstParagraph(contentRoot))),
		PublishedAt:      findPublishedAt(postNode, postNode),
		ImageCollections: findImageCollections(contentRoot, sourceURL),
	}
}

func findCandidateImageGroups(root *html.Node) []*html.Node {
	groups := make([]*html.Node, 0)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if isImageGroupNode(node) {
			groups = append(groups, node)
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return groups
}

func isImageGroupNode(node *html.Node) bool {
	if node == nil || node.Type != html.ElementNode {
		return false
	}
	if descendantCount(node, "img") == 0 {
		return false
	}
	if node.Data == "figure" {
		return true
	}

	classes := " " + nodeClass(node) + " "
	for _, hint := range []string{
		" gallery ",
		" wp-block-gallery ",
		" blocks-gallery-grid ",
		" slider ",
		" carousel ",
		" hero ",
		" slideshow ",
	} {
		if strings.Contains(classes, hint) {
			return true
		}
	}

	return false
}

func imagesFromNode(root *html.Node, baseURL string, seen map[string]struct{}) []catalog.ImageReference {
	images := make([]catalog.ImageReference, 0)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "img" {
			image := imageFromNode(node, baseURL)
			if strings.TrimSpace(image.SourceURL) != "" && isLikelyContentImage(node, image.SourceURL) {
				if _, exists := seen[image.SourceURL]; !exists {
					seen[image.SourceURL] = struct{}{}
					images = append(images, image)
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return images
}

func isLikelyContentImage(node *html.Node, sourceURL string) bool {
	lowerURL := strings.ToLower(sourceURL)
	if strings.Contains(lowerURL, "gravatar.com") ||
		strings.Contains(lowerURL, "avatar") ||
		strings.Contains(lowerURL, "emoji") ||
		strings.Contains(lowerURL, "smiley") ||
		strings.Contains(lowerURL, "sponsorlogga") ||
		strings.Contains(lowerURL, "leader-eu-logga") ||
		strings.Contains(lowerURL, "wp-header-image") {
		return false
	}

	classes := " " + strings.ToLower(nodeClass(node)) + " "
	if strings.Contains(classes, " avatar ") || strings.Contains(classes, " emoji ") || strings.Contains(classes, " smiley ") {
		return false
	}

	if width := attr(node, "width"); width != "" && (width == "32" || width == "40" || width == "48") {
		if height := attr(node, "height"); height == width {
			return false
		}
	}

	return true
}

func imageFromNode(node *html.Node, baseURL string) catalog.ImageReference {
	src := resolveURL(baseURL, preferredImageSource(node, baseURL))
	return catalog.ImageReference{
		ID:        "image-" + slugify(src),
		Title:     firstNonEmpty(attr(node, "title"), attr(node, "data-image-title")),
		SourceURL: src,
		AltText:   attr(node, "alt"),
		Caption:   firstNonEmpty(findAncestorCaption(node), findSiblingCaption(node)),
	}
}

func preferredImageSource(node *html.Node, baseURL string) string {
	if node == nil {
		return ""
	}
	if parent := node.Parent; parent != nil && parent.Type == html.ElementNode && parent.Data == "a" {
		href := resolveURL(baseURL, attr(parent, "href"))
		if looksLikeImageURL(href) {
			return href
		}
	}
	return attr(node, "src")
}

func findAncestorCaption(node *html.Node) string {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Type == html.ElementNode && current.Data == "figure" {
			for child := current.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "figcaption" {
					return normalizeWhitespace(textContent(child))
				}
			}
		}
	}
	return ""
}

func findSiblingCaption(node *html.Node) string {
	if node.Parent == nil {
		return ""
	}
	for sibling := node.Parent.FirstChild; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type != html.ElementNode {
			continue
		}
		if sibling.Data == "figcaption" || strings.Contains(" "+nodeClass(sibling)+" ", " caption ") {
			return normalizeWhitespace(textContent(sibling))
		}
	}
	return ""
}

func imageCollectionTitle(node *html.Node) string {
	if title := firstNonEmpty(
		attr(node, "aria-label"),
		findHeading(node, "h2"),
		findHeading(node, "h3"),
		findHeading(node, "h4"),
		findFigureCaption(node),
	); title != "" {
		return title
	}

	classes := nodeClass(node)
	switch {
	case strings.Contains(classes, "hero"):
		return "Hero Images"
	case strings.Contains(classes, "gallery"):
		return "Gallery Images"
	case strings.Contains(classes, "slider"), strings.Contains(classes, "carousel"):
		return "Carousel Images"
	default:
		return "Inline Images"
	}
}

func imageCollectionKind(node *html.Node, imageCount int) string {
	classes := nodeClass(node)
	switch {
	case strings.Contains(classes, "hero"):
		return "hero"
	case strings.Contains(classes, "gallery"):
		return "gallery"
	case strings.Contains(classes, "slider"), strings.Contains(classes, "carousel"), imageCount > 1:
		return "carousel"
	default:
		return "inline"
	}
}

func fallbackImageCollectionTitle(existingCollectionCount int, imageCount int) string {
	if existingCollectionCount == 0 && imageCount > 1 {
		return "Page Images"
	}
	if imageCount == 1 {
		return "Inline Image"
	}
	return "Inline Images"
}

func fallbackImageCollectionKind(imageCount int) string {
	if imageCount > 1 {
		return "gallery"
	}
	return "inline"
}

func findFigureCaption(node *html.Node) string {
	if node == nil {
		return ""
	}
	var caption string
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if caption != "" {
			return
		}
		if current.Type == html.ElementNode && current.Data == "figcaption" {
			caption = normalizeWhitespace(textContent(current))
			return
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return caption
}

func nodeClass(node *html.Node) string {
	return strings.TrimSpace(attr(node, "class"))
}

func hasClass(node *html.Node, wanted string) bool {
	classes := strings.Fields(nodeClass(node))
	for _, className := range classes {
		if className == wanted {
			return true
		}
	}
	return false
}

func findDescendantByClass(root *html.Node, className string) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found != nil {
			return
		}
		if hasClass(node, className) {
			found = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return found
}

func findFirstElement(root *html.Node, tag string) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found != nil {
			return
		}
		if node.Type == html.ElementNode && node.Data == tag {
			found = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return found
}

func descendantCount(root *html.Node, tag string) int {
	count := 0
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == tag {
			count++
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return count
}

func attr(node *html.Node, key string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == key {
			return strings.TrimSpace(attribute.Val)
		}
	}
	return ""
}

func textContent(node *html.Node) string {
	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return builder.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func resolveURL(baseURL string, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return ref
	}
	relative, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	resolved := base.ResolveReference(relative)
	if resolved.Hostname() == base.Hostname() && resolved.Scheme == "http" {
		resolved.Scheme = "https"
	}
	return resolved.String()
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func slugify(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(
		"https://", "",
		"http://", "",
		"/", "-",
		"?", "-",
		"&", "-",
		"=", "-",
		"_", "-",
		".", "-",
		":", "-",
	)
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	value = normalizeWhitespace(value)
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func looksLikeImageURL(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasSuffix(value, ".jpg") ||
		strings.HasSuffix(value, ".jpeg") ||
		strings.HasSuffix(value, ".png") ||
		strings.HasSuffix(value, ".gif") ||
		strings.HasSuffix(value, ".webp")
}

func slugFromURL(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if postID := parsed.Query().Get("p"); postID != "" {
		return "post-" + postID
	}
	trimmed := strings.Trim(parsed.Path, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	return slugify(parts[len(parts)-1])
}
