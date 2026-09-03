package main

import "testing"

func TestParseSemanticVersionAndComparison(t *testing.T) {
	current, err := parseSemanticVersion("v0.1.9")
	if err != nil {
		t.Fatal(err)
	}
	latest, err := parseSemanticVersion("0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !latest.newerThan(current) || current.newerThan(latest) {
		t.Fatal("comparação de versões incorreta")
	}
	for _, invalid := range []string{"1", "1.2", "1.2.3-beta", "1.a.3"} {
		if _, err := parseSemanticVersion(invalid); err == nil {
			t.Fatalf("versão inválida foi aceita: %s", invalid)
		}
	}
}

func TestReleaseAssetForPlatform(t *testing.T) {
	release := githubRelease{Assets: []releaseAsset{
		{Name: "highway_0.2.0-1_amd64.deb"},
		{Name: "Highway-0.2.0-macos-arm64.dmg"},
		{Name: "SHA256SUMS"},
	}}
	asset, err := releaseAssetForPlatform(release, "linux", "amd64")
	if err != nil || asset.Name != "highway_0.2.0-1_amd64.deb" {
		t.Fatalf("asset Linux incorreto: %#v, %v", asset, err)
	}
	asset, err = releaseAssetForPlatform(release, "darwin", "arm64")
	if err != nil || asset.Name != "Highway-0.2.0-macos-arm64.dmg" {
		t.Fatalf("asset macOS incorreto: %#v, %v", asset, err)
	}
	if _, err := releaseAssetForPlatform(release, "darwin", "amd64"); err == nil {
		t.Fatal("asset ausente foi aceito")
	}
}

func TestChecksumForAsset(t *testing.T) {
	checksum := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	content := checksum + "  highway_0.2.0-1_amd64.deb\n"
	got, err := checksumForAsset(content, "highway_0.2.0-1_amd64.deb")
	if err != nil || got != checksum {
		t.Fatalf("checksum incorreto: %q, %v", got, err)
	}
	if _, err := checksumForAsset(content, "ausente.deb"); err == nil {
		t.Fatal("checksum ausente foi aceito")
	}
}
