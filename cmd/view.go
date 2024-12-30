package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/cobra"
)

var supportedExtensions = []string{".jpg", ".jpeg", ".png"}

var viewCmd = &cobra.Command{
	Use:   "view [ファイルまたはディレクトリ]",
	Short: "EXIF情報を表示",
	Long:  "画像ファイルまたはディレクトリ内の画像ファイルのEXIF情報を表示します。",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		files := collectImageFiles(path)

		if len(files) == 0 {
			log.Fatalf("有効な画像ファイルが見つかりません: %s", path)
		}

		for _, file := range files {
			processFile(file)
			fmt.Println("--------------------")
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

func collectImageFiles(path string) []string {
	var files []string

	info, err := os.Stat(path)
	if err != nil {
		log.Fatalf("パスが無効です: %s (%v)", path, err)
	}

	// ファイルかどうかをチェック
	if !info.IsDir() {
		if isSupportedImage(path) {
			files = append(files, path)
		}
		return files
	}

	// ディレクトリなら再帰的に探索
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && isSupportedImage(p) {
			files = append(files, p)
		}
		return nil
	})

	if err != nil {
		log.Printf("ディレクトリの探索中にエラーが発生しました: %s (%v)", path, err)
	}
	return files
}

func isSupportedImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range supportedExtensions {
		if ext == supported {
			return true
		}
	}
	return false
}

func processFile(file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Printf("ファイルを開けません: %s (%v)", file, err)
		return
	}
	defer f.Close()

	fmt.Printf("ファイル: %s\n", file)

	meta, err := exif.Decode(f)
	if err != nil {
		log.Printf("EXIF情報を取得できません: %s (%v)", file, err)
		return
	}

	printExifMinimal(meta)
}

func printExifMinimal(meta *exif.Exif) {
	if date, err := meta.DateTime(); err == nil {
		fmt.Printf("撮影日: %s\n", date)
	}
	if model, err := meta.Get(exif.Model); err == nil {
		fmt.Printf("カメラモデル: %s\n", model.String())
	}
	if make, err := meta.Get(exif.Make); err == nil {
		fmt.Printf("メーカー: %s\n", make.String())
	}

	// GPS情報の確認
	lat, lon, err := meta.LatLong()
	if err == nil {
		fmt.Printf("GPS情報: 緯度 %.6f, 経度 %.6f\n", lat, lon)
	} else {
		fmt.Printf("GPS情報: 含まれていません\n")
	}
}
