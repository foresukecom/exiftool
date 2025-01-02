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

	"github.com/spf13/cobra"
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
	// メモリ上で画像をデコード
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("画像のデコードに失敗しました")
	}

	// JPEG形式でなければエラー
	if format != "jpeg" {
		return nil, errors.New("JPEG形式の画像のみサポートされています")
	}

	// EXIFデータを削除してバッファにエンコード
	var buffer bytes.Buffer
	err = jpeg.Encode(&buffer, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, errors.New("画像のエンコードに失敗しました")
	}

	return buffer.Bytes(), nil
}
