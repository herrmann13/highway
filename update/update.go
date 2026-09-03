package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"highway/version"
)

const (
	githubRepository = "herrmann13/highway"
	githubAPIBase    = "https://api.github.com/repos/" + githubRepository + "/releases/latest"
	updateTimeout    = 20 * time.Second
)

type GithubRelease struct {
	TagName string         `json:"tag_name"`
	HTMLURL string         `json:"html_url"`
	Body    string         `json:"body"`
	Assets  []ReleaseAsset `json:"assets"`
}

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type SemanticVersion struct {
	major int
	minor int
	patch int
}

func FetchLatestRelease(ctx context.Context) (GithubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIBase, nil)
	if err != nil {
		return GithubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Highway/"+version.AppVersion)

	client := &http.Client{Timeout: updateTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return GithubRelease{}, fmt.Errorf("não foi possível consultar atualizações: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return GithubRelease{}, fmt.Errorf("GitHub retornou %s ao consultar atualizações", resp.Status)
	}

	var release GithubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2*1024*1024)).Decode(&release); err != nil {
		return GithubRelease{}, fmt.Errorf("resposta de atualização inválida: %w", err)
	}
	if _, err := ParseSemanticVersion(release.TagName); err != nil {
		return GithubRelease{}, fmt.Errorf("a release do GitHub tem uma versão inválida: %w", err)
	}
	return release, nil
}

func ParseSemanticVersion(raw string) (SemanticVersion, error) {
	raw = strings.TrimPrefix(strings.TrimSpace(raw), "v")
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return SemanticVersion{}, fmt.Errorf("%q", raw)
	}
	version := SemanticVersion{}
	values := []*int{&version.major, &version.minor, &version.patch}
	for i, part := range parts {
		if part == "" {
			return SemanticVersion{}, fmt.Errorf("%q", raw)
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return SemanticVersion{}, fmt.Errorf("%q", raw)
			}
		}
		var value int
		if _, err := fmt.Sscanf(part, "%d", &value); err != nil {
			return SemanticVersion{}, fmt.Errorf("%q", raw)
		}
		*values[i] = value
	}
	return version, nil
}

func (v SemanticVersion) NewerThan(other SemanticVersion) bool {
	if v.major != other.major {
		return v.major > other.major
	}
	if v.minor != other.minor {
		return v.minor > other.minor
	}
	return v.patch > other.patch
}

func ReleaseAssetForPlatform(release GithubRelease, goos, goarch string) (ReleaseAsset, error) {
	var suffix string
	switch goos {
	case "linux":
		suffix = "_" + goarch + ".deb"
	case "darwin":
		suffix = "-macos-" + goarch + ".dmg"
	default:
		return ReleaseAsset{}, fmt.Errorf("atualização não suportada em %s", goos)
	}
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, suffix) {
			return asset, nil
		}
	}
	return ReleaseAsset{}, fmt.Errorf("a release não possui instalador para %s/%s", goos, goarch)
}

func checksumAsset(release GithubRelease) (ReleaseAsset, error) {
	for _, asset := range release.Assets {
		if asset.Name == "SHA256SUMS" {
			return asset, nil
		}
	}
	return ReleaseAsset{}, fmt.Errorf("a release não possui o arquivo SHA256SUMS")
}

func DownloadAsset(ctx context.Context, asset ReleaseAsset) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Highway/"+version.AppVersion)
	client := &http.Client{Timeout: updateTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("não foi possível baixar %s: %w", asset.Name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download de %s falhou: %s", asset.Name, resp.Status)
	}

	file, err := os.CreateTemp("", "highway-update-*"+filepath.Ext(asset.Name))
	if err != nil {
		return "", err
	}
	path := file.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	if _, err = io.Copy(file, io.LimitReader(resp.Body, 1024*1024*1024)); err != nil {
		file.Close()
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	return path, nil
}

func VerifyAssetChecksum(ctx context.Context, release GithubRelease, asset ReleaseAsset, path string) error {
	checksums, err := checksumAsset(release)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksums.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Highway/"+version.AppVersion)
	resp, err := (&http.Client{Timeout: updateTimeout}).Do(req)
	if err != nil {
		return fmt.Errorf("não foi possível baixar os checksums: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download de SHA256SUMS falhou: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return err
	}
	expected, err := ChecksumForAsset(string(data), asset.Name)
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	if !strings.EqualFold(expected, hex.EncodeToString(hash.Sum(nil))) {
		return fmt.Errorf("checksum inválido para %s", asset.Name)
	}
	return nil
}

func ChecksumForAsset(content, name string) (string, error) {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[1] != name {
			continue
		}
		if len(fields[0]) != sha256.Size*2 {
			break
		}
		if _, err := hex.DecodeString(fields[0]); err == nil {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksum não encontrado para %s", name)
}
