package main

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestThemeTooltipDelay(t *testing.T) {
	for _, dark := range []bool{false, true} {
		if got := lodestoneTheme(dark).TooltipDelay; got != 400*time.Millisecond {
			t.Fatalf("dark %v delay %v", dark, got)
		}
	}
}

func TestThemeChromeIsNeutral(t *testing.T) {
	// The dark look is kero's paper: tint 0, so no hue may leak into the
	// chrome. The light look is mist, where the blue-gray wash is the
	// design; the blocks alone are allowed the bright degrees of hue.
	tt := lodestoneTheme(true)
	for name, c := range map[string]ui.Color{
		"background":    tt.Background,
		"surface":       tt.Surface,
		"surface hover": tt.SurfaceHover,
		"text":          tt.Text,
		"text muted":    tt.TextMuted,
	} {
		lo, hi := min3(c.R, c.G, c.B), max3(c.R, c.G, c.B)
		if hi-lo > 6 {
			t.Fatalf("dark %s %v carries hue: %d channels apart", name, c, hi-lo)
		}
	}
	tt = lodestoneTheme(false)
	if l := luma(tt.Background); l < 225 {
		t.Fatalf("light background not paper: luma %.0f", l)
	}
	if l := luma(tt.Text); l > 90 {
		t.Fatalf("light text not ink: luma %.0f", l)
	}
	lo, hi := min3(tt.Accent.R, tt.Accent.G, tt.Accent.B), max3(tt.Accent.R, tt.Accent.G, tt.Accent.B)
	if hi-lo < 30 {
		t.Fatalf("light accent not the mist blue: %d channels apart", hi-lo)
	}
}

func min3(a, b, c uint8) uint8 {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

func max3(a, b, c uint8) uint8 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

func TestBlockGradientVariesAndShines(t *testing.T) {
	a := blockGradient("Hub", false, 0)
	b := blockGradient("Hub", false, 1)
	if len(a.Stops) < 2 || len(b.Stops) < 2 {
		t.Fatalf("stops %d %d", len(a.Stops), len(b.Stops))
	}
	if a.Stops[0].Color == b.Stops[0].Color && a.Stops[len(a.Stops)-1].Color == b.Stops[len(b.Stops)-1].Color {
		t.Fatal("neighboring hub blocks share a fill")
	}
	if blockGradient("Runners", false, 0).Stops[1].Color == a.Stops[1].Color {
		t.Fatal("runners and hub share a mid color")
	}
	rest := shadeGradient(a, -1, false)
	dim := shadeGradient(a, 0, false)
	hot := shadeGradient(a, 1, false)
	muted := shadeGradient(a, -1, true)
	if luma(hot.Stops[0].Color) <= luma(dim.Stops[0].Color) {
		t.Fatalf("shine %v is not lighter than dim %v", hot.Stops[0].Color, dim.Stops[0].Color)
	}
	if luma(muted.Stops[0].Color) >= luma(rest.Stops[0].Color) {
		t.Fatalf("unselected %v is not darker than rest %v", muted.Stops[0].Color, rest.Stops[0].Color)
	}
}

func TestTipTitleUsesFileName(t *testing.T) {
	if got := tipTitle("Qwen/Qwen3-0.6B", "/home/leilei/.cache/huggingface/hub/model.safetensors"); got != "model.safetensors · Qwen/Qwen3-0.6B" {
		t.Fatalf("title %q", got)
	}
	if got := tipTitle("hub", "/cache/hub"); got != "hub" {
		t.Fatalf("same base %q", got)
	}
	if got := tipTitle("", "/cache/weights.bin"); got != "weights.bin" {
		t.Fatalf("base only %q", got)
	}
}

func TestTipFollowsPointer(t *testing.T) {
	self, dx, dy := tipFollow(40, 30, 800, 500, true)
	if self != ui.AnchorTopLeft || dx != 54 || dy != 46 {
		t.Fatalf("below right: %v %v %v", self, dx, dy)
	}
	self, dx, dy = tipFollow(700, 30, 800, 500, true)
	if self != ui.AnchorTopRight || dx != 686 || dy != 46 {
		t.Fatalf("below left: %v %v %v", self, dx, dy)
	}
	self, dx, dy = tipFollow(40, 460, 800, 500, true)
	if self != ui.AnchorBottomLeft || dx != 54 || dy != 448 {
		t.Fatalf("above right: %v %v %v", self, dx, dy)
	}
	self, dx, dy = tipFollow(700, 460, 800, 500, true)
	if self != ui.AnchorBottomRight || dx != 686 || dy != 448 {
		t.Fatalf("above left: %v %v %v", self, dx, dy)
	}
}

func luma(c ui.Color) float32 {
	return 0.2126*float32(c.R) + 0.7152*float32(c.G) + 0.0722*float32(c.B)
}
