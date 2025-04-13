package test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
)

func RunTests(command string) {
	if command == "" {
		fmt.Println("使用方法: [generate|test]")
		fmt.Println("  generate: テスト画像を生成")
		fmt.Println("  test: 向き情報の修正をテスト")
		return
	}

	switch command {
	case "generate":
		fmt.Println("テスト画像を生成します...")
		generateTestImages()
	case "test":
		fmt.Println("向き情報の修正をテストします...")
		testOrientationFix()
	default:
		fmt.Printf("不明なコマンド: %s\n", command)
		fmt.Println("使用方法: [generate|test]")
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
	
	jpegData := buf.Bytes()
	exifData := createExifData(orientation)
	
	result := make([]byte, 0, len(jpegData) + len(exifData) + 10)
	result = append(result, jpegData[0:2]...) // SOIマーカー
	
	result = append(result, 0xFF, 0xE1)
	
	exifLength := len(exifData) + 2 // 長さ自体の2バイト
	result = append(result, byte((exifLength>>8)&0xFF), byte(exifLength&0xFF))
	
	result = append(result, exifData...)
	
	result = append(result, jpegData[2:]...)
	
	if err := os.WriteFile(filename, result, 0644); err != nil {
		return fmt.Errorf("ファイルの保存に失敗: %v", err)
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

func createExifData(orientation int) []byte {
	exifHeader := []byte{
		'E', 'x', 'i', 'f', 0x00, 0x00,
		'M', 'M', 0x00, 0x2A, // バイトオーダー (MM = big endian)
		0x00, 0x00, 0x00, 0x08, // IFDオフセット
	}
	
	ifdCount := []byte{0x00, 0x01}
	
	orientationTag := []byte{
		0x01, 0x12, // タグ番号 (Orientation = 0x0112)
		0x00, 0x03, // データ型 (SHORT = 3)
		0x00, 0x00, 0x00, 0x01, // カウント (1)
		0x00, 0x00, 0x00, byte(orientation), // 値 (orientationの値)
	}
	
	ifdEnd := []byte{0x00, 0x00, 0x00, 0x00}
	
	exifData := make([]byte, 0, len(exifHeader)+len(ifdCount)+len(orientationTag)+len(ifdEnd))
	exifData = append(exifData, exifHeader...)
	exifData = append(exifData, ifdCount...)
	exifData = append(exifData, orientationTag...)
	exifData = append(exifData, ifdEnd...)
	
	return exifData
}

func getOrientation(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	buf := make([]byte, 4)
	if _, err := file.Read(buf); err != nil {
		return 0, err
	}
	
	if buf[0] != 0xFF || buf[1] != 0xD8 {
		return 0, fmt.Errorf("JPEGファイルではありません")
	}
	
	for {
		if _, err := file.Read(buf[:2]); err != nil {
			return 0, err
		}
		
		if buf[0] != 0xFF {
			continue
		}
		
		if buf[1] == 0xE1 {
			if _, err := file.Read(buf[:2]); err != nil {
				return 0, err
			}
			length := int(buf[0])<<8 | int(buf[1])
			
			exifHeader := make([]byte, 6)
			if _, err := file.Read(exifHeader); err != nil {
				return 0, err
			}
			
			if string(exifHeader[:4]) == "Exif" {
				exifData := make([]byte, length-8) // 長さ - 2バイト(長さ自体) - 6バイト(Exifヘッダー)
				if _, err := file.Read(exifData); err != nil {
					return 0, err
				}
				
				var byteOrder binary.ByteOrder
				if exifData[0] == 'I' && exifData[1] == 'I' {
					byteOrder = binary.LittleEndian
				} else if exifData[0] == 'M' && exifData[1] == 'M' {
					byteOrder = binary.BigEndian
				} else {
					return 0, fmt.Errorf("不明なバイトオーダー")
				}
				
				ifdOffset := byteOrder.Uint32(exifData[4:8])
				
				entryCount := byteOrder.Uint16(exifData[ifdOffset:ifdOffset+2])
				
				for i := uint16(0); i < entryCount; i++ {
					entryOffset := ifdOffset + 2 + uint32(i)*12
					tag := byteOrder.Uint16(exifData[entryOffset : entryOffset+2])
					
					if tag == 0x0112 {
						valueOffset := entryOffset + 8
						return int(byteOrder.Uint16(exifData[valueOffset : valueOffset+2])), nil
					}
				}
			}
		}
		
		if buf[1] == 0xDA {
			break
		}
		
		if _, err := file.Read(buf[:2]); err != nil {
			return 0, err
		}
		length := int(buf[0])<<8 | int(buf[1])
		if _, err := file.Seek(int64(length-2), os.SEEK_CUR); err != nil {
			return 0, err
		}
	}
	
	return 1, nil
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
