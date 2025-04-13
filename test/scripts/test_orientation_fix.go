package tester

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rwcarlsen/goexif/exif"
)

func TestOrientationFix() {
	imagesDir := "../images"
	
	resultsDir := "../results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		fmt.Printf("結果ディレクトリの作成に失敗しました: %v\n", err)
		return
	}
	
	exiftoolBin := "../../exiftool"
	
	fmt.Println("exiftoolをビルドしています...")
	cmd := exec.Command("go", "build", "-o", exiftoolBin)
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
		cmd := exec.Command(exiftoolBin, "remove", resultFile)
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
