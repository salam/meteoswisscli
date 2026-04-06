package output

import (
	"image"
	"image/color"
	"testing"
)

func TestPixelToBlock(t *testing.T) {
	tests := []struct {
		brightness uint8
		want       rune
	}{
		{0, ' '},
		{64, '░'},
		{128, '▒'},
		{192, '▓'},
		{255, '█'},
	}
	for _, tt := range tests {
		got := pixelToBlock(tt.brightness)
		if got != tt.want {
			t.Errorf("pixelToBlock(%d) = %c, want %c", tt.brightness, got, tt.want)
		}
	}
}

func TestRenderASCII(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetGray(x, y, color.Gray{Y: 200})
		}
	}
	result := RenderASCII(img, 4)
	if len(result) == 0 {
		t.Error("RenderASCII should produce output")
	}
}
