package wordpress

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lorudden/hemsida/internal/content/catalog"
)

const legacyMediaStoragePrefix = "legacy"

type legacyMediaStore struct {
	client HTTPClient
	root   string
}

func newLegacyMediaStore(client HTTPClient, root string) *legacyMediaStore {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}

	return &legacyMediaStore{
		client: client,
		root:   root,
	}
}

func (s *legacyMediaStore) Download(ctx context.Context, image catalog.ImageReference) (catalog.ImageReference, error) {
	sourceURL := strings.TrimSpace(image.SourceURL)
	if sourceURL == "" {
		return image, nil
	}

	urlExt := imageExtensionFromURL(sourceURL)
	if urlExt != "" {
		storagePath := legacyMediaStoragePath(sourceURL, urlExt)
		mimeType, err := s.ensureExistingAsset(storagePath)
		if err != nil {
			return image, err
		}
		if mimeType != "" {
			image.StoragePath = storagePath
			image.MIMEType = mimeType
			return image, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return image, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return image, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return image, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return image, fmt.Errorf("read legacy image %q: %w", sourceURL, err)
	}

	mimeType := normalizeMIMEType(resp.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = http.DetectContentType(payload)
	}

	ext := urlExt
	if ext == "" {
		ext = imageExtensionFromMIMEType(mimeType)
	}
	if ext == "" {
		ext = ".img"
	}

	storagePath := legacyMediaStoragePath(sourceURL, ext)
	absolutePath := filepath.Join(s.root, filepath.FromSlash(storagePath))

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return image, fmt.Errorf("prepare legacy media directory for %q: %w", sourceURL, err)
	}

	if err := os.WriteFile(absolutePath, payload, 0o644); err != nil {
		return image, fmt.Errorf("write legacy image %q: %w", sourceURL, err)
	}

	image.StoragePath = storagePath
	image.MIMEType = mimeType
	return image, nil
}

func (s *legacyMediaStore) ensureExistingAsset(storagePath string) (string, error) {
	absolutePath := filepath.Join(s.root, filepath.FromSlash(storagePath))
	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("stat legacy media %q: %w", absolutePath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("legacy media path %q is a directory", absolutePath)
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return "", fmt.Errorf("open legacy media %q: %w", absolutePath, err)
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read legacy media %q: %w", absolutePath, err)
	}

	return http.DetectContentType(header[:n]), nil
}

func legacyMediaStoragePath(sourceURL string, ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	sum := sha256.Sum256([]byte(strings.TrimSpace(sourceURL)))
	name := hex.EncodeToString(sum[:12]) + ext
	return path.Join(legacyMediaStoragePrefix, name)
}

func imageExtensionFromURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}

	ext := strings.ToLower(path.Ext(parsed.Path))
	if isKnownImageExtension(ext) {
		return ext
	}

	return ""
}

func imageExtensionFromMIMEType(value string) string {
	value = normalizeMIMEType(value)
	if value == "" {
		return ""
	}

	if exts, err := mime.ExtensionsByType(value); err == nil {
		for _, ext := range exts {
			if isKnownImageExtension(ext) {
				return ext
			}
		}
	}

	switch value {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/avif":
		return ".avif"
	default:
		return ""
	}
}

func normalizeMIMEType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	mediaType, _, err := mime.ParseMediaType(value)
	if err == nil {
		return strings.ToLower(strings.TrimSpace(mediaType))
	}

	return strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
}

func isKnownImageExtension(ext string) bool {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".avif":
		return true
	default:
		return false
	}
}
