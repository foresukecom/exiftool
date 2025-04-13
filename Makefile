.PHONY: build test generate-test-images run-tests clean

# デフォルトターゲット
all: build

# ビルド
build:
	go build -o exiftool

# テスト画像の生成
generate-test-images:
	cd test/scripts && go run main_runner.go generate

# テストの実行
run-tests: build generate-test-images
	cd test/scripts && go run main_runner.go test

# テスト環境のセットアップと実行（一括実行）
test: build generate-test-images run-tests

# クリーンアップ
clean:
	rm -f exiftool
	rm -rf test/images/*
	rm -rf test/results/*

# ヘルプ
help:
	@echo "使用可能なコマンド:"
	@echo "  make build              - exiftoolをビルド"
	@echo "  make generate-test-images - テスト画像を生成"
	@echo "  make run-tests          - テストを実行"
	@echo "  make test               - テスト画像の生成とテストを一括実行"
	@echo "  make clean              - ビルドファイルとテスト結果を削除"
