# SOJ

Support Online Judge

これはプログラミングコンテストの作問をサポートするプログラムです

# 準備

まずgoをインストールします

```
go get github.com/yosupo06/soj
```

します

これで実行ファイルが勝手にコンパイルされて$GOPATH/binに入るはずです

# 使い方

| フォルダ |                       中身                       |
|----------|--------------------------------------------------|
| ./       | tomlファイル                                     |
| ./maker  | テストケース作成プログラム, ベリファイプログラム |
| ./answer | 解答プログラム                                   |

という感じに置いて `./` で

```
soj -toml=tomlfile.toml
```

します

テストケース作成プログラムは

```
./a.out	--seed={そのプログラムを実行するのは何回目か} --hash={適当な値}
```

みたいな感じで実行されるので標準出力にテストケースを出すプログラムを書きます
hashは32ビット符号なし整数に収まります

解答プログラムは普通に標準入力から入力して標準出力に出すものを用意します

ベリファイは標準入力から読み込んで正しかったらreturn 0, 違ったらreturn 1

tomlファイルは

```
answer = "ac.cpp"
verify = "verify.d"
timeLimit = 5.0
[[tests]]
	Name = "smallrandom.d"
	Number = 10
[[tests]]
	Name = "bigrandom.d"
	Number = 5
[[tests]]
	Name = "maxrandom.d"
	Number = 2


[[answers]]
	Name = "ac2.cpp"
[[answers]]
	Name = "wa.cpp"
[[answers]]
	Name = "tle.cpp"
```

みたいな感じで

