package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

func main() {
	outputDir := "../images"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("ディレクトリの作成に失敗しました: %v\n", err)
		return
	}

	for orientation := 1; orientation <= 8; orientation++ {
		img := createTestImage(orientation)
		
		if err := saveImageWithOrientation(img, filepath.Join(outputDir, fmt.Sprintf("test_orientation_%d.jpg", orientation)), orientation); err != nil {
			fmt.Printf("画像の保存に失敗しました (orientation=%d): %v\n", orientation, err)
		} else {
			fmt.Printf("テスト画像を生成しました: test_orientation_%d.jpg\n", orientation)
		}
	}
}

func createTestImage(orientation int) image.Image {
	width, height := 300, 200
	
	if orientation == 5 || orientation == 6 || orientation == 7 || orientation == 8 {
		width, height = height, width
	}
	
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	
	for y := 0; y < 40; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	
	for y := height/2 - 30; y < height/2 + 30; y++ {
		for x := width/2 - 30; x < width/2 + 30; x++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
	
	return img
}

func saveImageWithOrientation(img image.Image, filename string, orientation int) error {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return fmt.Errorf("JPEGエンコードに失敗: %v", err)
	}
	
	e := exif.Exif{
		Tiff: tiff.Tiff{
			Dirs: []*tiff.Dir{
				{
					Fields: map[uint16]*tiff.Field{
						0x0112: tiff.NewField(0x0112, tiff.Short, 1, []uint16{uint16(orientation)}),
					},
				},
			},
		},
	}
	
	var exifBuf bytes.Buffer
	if err := e.Encode(&exifBuf); err != nil {
		return fmt.Errorf("EXIFエンコードに失敗: %v", err)
	}
	
	jpegData := buf.Bytes()
	exifData := exifBuf.Bytes()
	
	result := make([]byte, 0, len(jpegData) + len(exifData) + 10)
	result = append(result, jpegData[0:2]...) // SOIマーカー
	
	result = append(result, 0xFF, 0xE1)
	
	exifLength := len(exifData) + 2 + 6 // 長さ自体の2バイト + "Exif\0\0"の6バイト
	result = append(result, byte((exifLength>>8)&0xFF), byte(exifLength&0xFF))
	
	result = append(result, []byte("Exif\000\000")...)
	
	result = append(result, exifData...)
	
	result = append(result, jpegData[2:]...)
	
	return os.WriteFile(filename, result, 0644)
}
