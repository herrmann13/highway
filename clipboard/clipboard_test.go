package clipboard

import "testing"

func TestCurlClipboardDetector(t *testing.T) {
	detector := &CurlClipboardDetector{}
	if _, ok := detector.Detect("texto comum"); ok {
		t.Fatal("texto comum foi identificado como cURL")
	}
	command := "curl --url https://api.example.com/users"
	if got, ok := detector.Detect(command); !ok || got != command {
		t.Fatalf("cURL não foi detectado: %q, %t", got, ok)
	}
	if _, ok := detector.Detect(command); ok {
		t.Fatal("o mesmo cURL foi detectado duas vezes")
	}
	if _, ok := detector.Detect("outro texto"); ok {
		t.Fatal("texto comum foi identificado como cURL")
	}
	if _, ok := detector.Detect(command); !ok {
		t.Fatal("cURL copiado novamente não foi detectado")
	}
}
