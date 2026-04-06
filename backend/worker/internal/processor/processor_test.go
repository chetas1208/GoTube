package processor

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestThumbnailCandidateSeconds(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		want     []int
	}{
		{name: "zero duration", duration: 0, want: []int{1}},
		{name: "short duration", duration: 5, want: []int{1, 2, 3}},
		{name: "long duration", duration: 120, want: []int{12, 24, 42, 60, 78}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thumbnailCandidateSeconds(tt.duration)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("thumbnailCandidateSeconds(%d) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func TestScoreThumbnailPrefersDetailedMidExposureFrame(t *testing.T) {
	tempDir := t.TempDir()
	darkPath := filepath.Join(tempDir, "dark.jpg")
	detailedPath := filepath.Join(tempDir, "detailed.jpg")

	if err := writeJPEG(darkPath, solidImage(320, 180, color.RGBA{R: 15, G: 15, B: 15, A: 255})); err != nil {
		t.Fatalf("writeJPEG(dark) error = %v", err)
	}
	if err := writeJPEG(detailedPath, checkerboardImage(320, 180)); err != nil {
		t.Fatalf("writeJPEG(detailed) error = %v", err)
	}

	darkScore, err := scoreThumbnail(darkPath)
	if err != nil {
		t.Fatalf("scoreThumbnail(dark) error = %v", err)
	}
	detailedScore, err := scoreThumbnail(detailedPath)
	if err != nil {
		t.Fatalf("scoreThumbnail(detailed) error = %v", err)
	}

	if detailedScore <= darkScore {
		t.Fatalf("expected detailed frame score to be higher: detailed=%f dark=%f", detailedScore, darkScore)
	}
}

func writeJPEG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
}

func solidImage(width, height int, fill color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fill)
		}
	}
	return img
}

func checkerboardImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	blockSize := 16
	palette := []color.RGBA{
		{R: 70, G: 95, B: 125, A: 255},
		{R: 205, G: 195, B: 135, A: 255},
		{R: 150, G: 85, B: 70, A: 255},
		{R: 230, G: 220, B: 210, A: 255},
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := ((x / blockSize) + (y / blockSize)) % len(palette)
			img.Set(x, y, palette[index])
		}
	}

	return img
}
