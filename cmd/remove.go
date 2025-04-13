package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/cobra"
	"golang.org/x/image/draw"
)

var outputDir string

var removeCmd = &cobra.Command{
	Use:   "remove [ファイルまたはディレクトリ]",
	Short: "EXIF情報を削除",
	Long:  "画像ファイルまたはディレクトリ内の画像ファイルからEXIF情報を削除します。",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		files := collectImageFiles(path)

		if len(files) == 0 {
			log.Fatalf("有効な画像ファイルが見つかりません: %s", path)
		}

		for _, file := range files {
			if err := removeEXIF(file, outputDir); err != nil {
				log.Printf("エラー: %s (%v)", file, err)
			} else {
				fmt.Printf("EXIF情報を削除しました: %s\n", file)
			}
		}
	},
}

func init() {
	// コマンドにオプションを追加
	removeCmd.Flags().StringVarP(&outputDir, "output", "o", "", "出力先ディレクトリ (指定しない場合は上書き保存)")
	rootCmd.AddCommand(removeCmd)
}

func removeEXIF(file string, outputDir string) error {
	// ファイルを開く
	input, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("ファイルを開けません: %v", err)
	}
	defer input.Close()

	// ファイル全体をメモリにロード
	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, input); err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗: %v", err)
	}

	// EXIFデータを削除
	cleanedFile, err := removeExifData(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("EXIF情報の削除に失敗: %v", err)
	}

	// アウトプット先を決定
	outputPath := file // デフォルトは元ファイルを置き換え
	if outputDir != "" {
		// 出力先ディレクトリが指定されている場合、新しいパスを作成
		baseName := filepath.Base(file)
		outputPath = filepath.Join(outputDir, baseName)

		// ディレクトリが存在しない場合は作成
		if err = os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("出力先ディレクトリの作成に失敗: %v", err)
		}
	}

	// ファイルを保存
	if err = os.WriteFile(outputPath, cleanedFile, 0644); err != nil {
		return fmt.Errorf("ファイルの保存に失敗: %v", err)
	}

	return nil
}

func removeExifData(data []byte) ([]byte, error) {
	orientation := 1 // デフォルト値（通常の向き）
	
	x, err := exif.Decode(bytes.NewReader(data))
	if err == nil {
		if tag, err := x.Get(exif.Orientation); err == nil {
			if val, err := tag.Int(0); err == nil {
				orientation = int(val)
			}
		}
	}
	
	// メモリ上で画像をデコード
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("画像のデコードに失敗しました")
	}
	
	// JPEG形式でなければエラー
	if format != "jpeg" {
		return nil, errors.New("JPEG形式の画像のみサポートされています")
	}
	
	img = applyOrientation(img, orientation)
	
	// EXIFデータを削除してバッファにエンコード
	var buffer bytes.Buffer
	err = jpeg.Encode(&buffer, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, errors.New("画像のエンコードに失敗しました")
	}
	
	return buffer.Bytes(), nil
}
func applyOrientation(img image.Image, orientation int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	if orientation == 1 {
		return img
	}
	
	var dst *image.RGBA
	
	if orientation == 5 || orientation == 6 || orientation == 7 || orientation == 8 {
		dst = image.NewRGBA(image.Rect(0, 0, height, width))
	} else {
		dst = image.NewRGBA(image.Rect(0, 0, width, height))
	}
	
	switch orientation {
	case 2: // 水平方向に反転
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(width-x-1, y, img.At(x, y))
			}
		}
	case 3: // 180度回転
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(width-x-1, height-y-1, img.At(x, y))
			}
		}
	case 4: // 垂直方向に反転
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(x, height-y-1, img.At(x, y))
			}
		}
	case 5: // 270度回転して水平方向に反転
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(y, x, img.At(x, y))
			}
		}
	case 6: // 90度回転 (時計回り)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(height-y-1, x, img.At(x, y))
			}
		}
	case 7: // 90度回転して水平方向に反転
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(height-y-1, width-x-1, img.At(x, y))
			}
		}
	case 8: // 270度回転 (90度反時計回り)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(y, width-x-1, img.At(x, y))
			}
		}
	default:
		draw.Draw(dst, dst.Bounds(), img, bounds.Min, draw.Src)
	}
	
	return dst
}
