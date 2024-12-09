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

	"github.com/spf13/cobra"
)

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
			removeEXIF(file)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func removeEXIF(file string) {
	// ファイルを開く
	input, err := os.Open(file)
	if err != nil {
		log.Printf("ファイルを開けません: %s (%v)", file, err)
		return
	}
	defer input.Close()

	// ファイル全体をメモリにロード
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, input)
	if err != nil {
		log.Printf("ファイルの読み込みに失敗しました: %s (%v)", file, err)
		return
	}

	// EXIFデータを除去（再エンコード処理）
	cleanedFile, err := removeExifData(buffer.Bytes())
	if err != nil {
		log.Printf("EXIF情報の削除に失敗しました: %s (%v)", file, err)
		return
	}

	// 元のファイルに上書き保存
	err = os.WriteFile(file, cleanedFile, 0644)
	if err != nil {
		log.Printf("ファイルの保存に失敗しました: %s (%v)", file, err)
	} else {
		fmt.Printf("EXIF情報を削除しました: %s\n", file)
	}
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
