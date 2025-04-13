package main

import (
	"fmt"
	"os"
	
	"./generator"
	"./tester"
)

func main() {
	fmt.Println("テスト環境をセットアップします...")
	
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run main.go [generate|test]")
		fmt.Println("  generate: テスト画像を生成")
		fmt.Println("  test: 向き情報の修正をテスト")
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "generate":
		fmt.Println("テスト画像を生成します...")
		generator.GenerateTestImages()
	case "test":
		fmt.Println("向き情報の修正をテストします...")
		tester.TestOrientationFix()
	default:
		fmt.Printf("不明なコマンド: %s\n", command)
		fmt.Println("使用方法: go run main.go [generate|test]")
	}
}
