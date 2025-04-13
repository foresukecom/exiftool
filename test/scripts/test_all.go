package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/rwcarlsen/goexif/exif"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run test_all.go [generate|test]")
		fmt.Println("  generate: テスト画像を生成")
		fmt.Println("  test: 向き情報の修正をテスト")
		return
	}

	command := os.Args[1]

	switch command {
	case "generate":
		fmt.Println("テスト画像を生成します...")
		generateTestImages()
	case "test":
		fmt.Println("向き情報の修正をテストします...")
		testOrientationFix()
	default:
		fmt.Printf("不明なコマンド: %s\n", command)
		fmt.Println("使用方法: go run test_all.go [generate|test]")
	}
}

func generateTestImages() {
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
	
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("ファイルの保存に失敗: %v", err)
	}
	
	cmd := exec.Command("exiftool", "-Orientation="+strconv.Itoa(orientation), "-overwrite_original", filename)
	if err := cmd.Run(); err != nil {
		fmt.Printf("警告: exiftoolでの向き情報の設定に失敗しました: %v\n", err)
		fmt.Printf("テスト画像は生成されましたが、向き情報が設定されていない可能性があります。\n")
	}
	
	return nil
}

func testOrientationFix() {
	imagesDir := "../images"
	
	resultsDir := "../results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		fmt.Printf("結果ディレクトリの作成に失敗しました: %v\n", err)
		return
	}
	
	ourExiftoolBin := "../../exiftool"
	
	fmt.Println("exiftoolをビルドしています...")
	cmd := exec.Command("go", "build", "-o", ourExiftoolBin)
	cmd.Dir = "../.."
	if err := cmd.Run(); err != nil {
		fmt.Printf("exiftoolのビルドに失敗しました: %v\n", err)
		return
	}
	
	for i := 1; i <= 8; i++ {
		testFile := filepath.Join(imagesDir, fmt.Sprintf("test_orientation_%d.jpg", i))
		resultFile := filepath.Join(resultsDir, fmt.Sprintf("result_orientation_%d.jpg", i))
		
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			fmt.Printf("テスト画像が見つかりません: %s\n", testFile)
			continue
		}
		
		if err := copyFile(testFile, resultFile); err != nil {
			fmt.Printf("テスト画像のコピーに失敗しました: %v\n", err)
			continue
		}
		
		originalOrientation, err := getOrientation(testFile)
		if err != nil {
			fmt.Printf("元の画像の向き情報の取得に失敗しました: %v\n", err)
			continue
		}
		
		fmt.Printf("テスト %d: EXIFデータを削除しています...\n", i)
		cmd := exec.Command(ourExiftoolBin, "remove", resultFile)
		if err := cmd.Run(); err != nil {
			fmt.Printf("EXIFデータの削除に失敗しました: %v\n", err)
			continue
		}
		
		processedWidth, processedHeight, err := getImageDimensions(resultFile)
		if err != nil {
			fmt.Printf("処理後の画像サイズの取得に失敗しました: %v\n", err)
			continue
		}
		
		originalWidth, originalHeight, err := getImageDimensions(testFile)
		if err != nil {
			fmt.Printf("元の画像サイズの取得に失敗しました: %v\n", err)
			continue
		}
		
		fmt.Printf("テスト %d (Orientation=%d):\n", i, originalOrientation)
		fmt.Printf("  元の画像サイズ: %dx%d\n", originalWidth, originalHeight)
		fmt.Printf("  処理後の画像サイズ: %dx%d\n", processedWidth, processedHeight)
		
		if (originalOrientation >= 5 && originalOrientation <= 8) {
			if originalWidth == processedHeight && originalHeight == processedWidth {
				fmt.Printf("  結果: ✅ 成功（幅と高さが正しく入れ替わりました）\n")
			} else {
				fmt.Printf("  結果: ❌ 失敗（幅と高さが正しく入れ替わっていません）\n")
			}
		} else {
			if originalWidth == processedWidth && originalHeight == processedHeight {
				fmt.Printf("  結果: ✅ 成功（幅と高さが保持されています）\n")
			} else {
				fmt.Printf("  結果: ❌ 失敗（幅と高さが変わっています）\n")
			}
		}
		
		fmt.Println("  処理後の画像: ", resultFile)
		fmt.Println("--------------------")
	}
	
	fmt.Println("テスト完了！")
	fmt.Println("元の画像: ", imagesDir)
	fmt.Println("処理後の画像: ", resultsDir)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func getOrientation(filename string) (int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	
	x, err := exif.Decode(f)
	if err != nil {
		return 0, err
	}
	
	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 0, err
	}
	
	val, err := tag.Int(0)
	if err != nil {
		return 0, err
	}
	
	return int(val), nil
}

func getImageDimensions(filename string) (int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	
	img, _, err := image.Decode(file)
	if err != nil {
		return 0, 0, err
	}
	
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), nil
}
