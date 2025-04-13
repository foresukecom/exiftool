package main

import (
	"fmt"
	"os"
	
	"exiftool/test/scripts/test"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run main_runner.go [generate|test]")
		fmt.Println("  generate: テスト画像を生成")
		fmt.Println("  test: 向き情報の修正をテスト")
		return
	}

	command := os.Args[1]
	
	test.RunTests(command)
}
