package output

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

func pixelToBlock(brightness uint8) rune {
	switch {
	case brightness < 51:
		return ' '
	case brightness < 102:
		return '░'
	case brightness < 153:
		return '▒'
	case brightness < 204:
		return '▓'
	default:
		return '█'
	}
}

func RenderASCII(img image.Image, width int) string {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	if imgW == 0 || imgH == 0 {
		return ""
	}

	height := width * imgH / imgW / 2
	if height < 1 {
		height = 1
	}

	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := bounds.Min.X + x*imgW/width
			srcY := bounds.Min.Y + y*imgH/height
			r, g, b, _ := img.At(srcX, srcY).RGBA()
			brightness := uint8((r/256*299 + g/256*587 + b/256*114) / 1000)
			sb.WriteRune(pixelToBlock(brightness))
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

func ASCIIMap(imageURL string, width int) error {
	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	if width <= 0 {
		width = 80
	}
	fmt.Print(RenderASCII(img, width))
	return nil
}

func SaveImage(imageURL string, path string) error {
	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
