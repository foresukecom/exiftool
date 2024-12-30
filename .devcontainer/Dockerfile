FROM ubuntu:latest

# 必要なパッケージをインストール
RUN apt-get update && apt-get install -y \
    curl git vim zsh sudo locales tree software-properties-common make build-essential

# タイムゾーンをJSTに設定
ENV TZ=Asia/Tokyo
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# ロケールを日本語に設定
RUN locale-gen ja_JP.UTF-8
ENV LANG ja_JP.UTF-8
ENV LANGUAGE ja_JP:ja
ENV LC_ALL ja_JP.UTF-8

# Golangのセットアップ
RUN add-apt-repository ppa:longsleep/golang-backports
RUN apt-get install -y golang-go

# Goの環境変数を設定
# Golangのインストール後にGOROOTを設定
ENV GOROOT /usr/lib/go
ENV GOPATH /workspace/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH
ENV GO111MODULE on

RUN go install github.com/spf13/cobra-cli@latest

# GoReleaserのインストール
RUN echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' \
    | tee /etc/apt/sources.list.d/goreleaser.list \
 && apt-get update \
 && apt-get install -y goreleaser
 
# 作業ディレクトリを設定
WORKDIR /workspace

# デフォルトコマンドを設定
CMD [ "/bin/zsh" ]