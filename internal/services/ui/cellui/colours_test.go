package cellui

import (
	"gosheet/internal/utils"
	"testing"
)

func TestSwatchColorDefaultMatchesCellDefaults(t *testing.T) {
	if got, want := swatchColor("", true), (utils.ColorRGB{255, 255, 255}); got != want {
		t.Fatalf("text default = %v, want %v", got, want)
	}
	if got, want := swatchColor("", false), (utils.ColorRGB{0, 0, 0}); got != want {
		t.Fatalf("background default = %v, want %v", got, want)
	}
}

func TestSwatchColorUsesPaletteHex(t *testing.T) {
	if got, want := swatchColor("#e06c75", true), (utils.ColorRGB{224, 108, 117}); got != want {
		t.Fatalf("palette color = %v, want %v", got, want)
	}
}
