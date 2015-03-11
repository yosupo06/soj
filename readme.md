# SOJ

Support Online Judge

これはプログラミングコンテストの作問をサポートするプログラムです

# 準備

まずgoをインストールします

```
go get github.com/yosupo06/soj
```

します

# 使い方

テストケース作成プログラム、解答プログラム、ベリファイ、tomlファイルを全部同じフォルダに全部突っ込んで

```
soj -toml=maketest.toml
```

します

テストケース作成プログラムは

```
./a.out	--seed=12342132
```

みたいな感じで実行されるので標準出力にテストケースを出します

解答プログラムは普通に標準入力から入力して標準出力に出す感じの

ベリファイは標準入力から読み込んで正しかったらreturn 0, 違ったらreturn 1

tomlファイルは

```
answer = "ac.cpp"
verify = "verify.d"
timeLimit = 5.0
[[tests]]
	Name = "smallrand.d"
	Number = 10
[[tests]]
	Name = "bigrand.d"
	Number = 5
[[tests]]
	Name = "maxrand.d"
	Number = 2


[[answers]]
	Name = "ac2.cpp"
[[answers]]
	Name = "wa.cpp"
[[answers]]
	Name = "tle.cpp"
```

みたいな感じで

