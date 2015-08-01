package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"os"
	fp "path/filepath"
)

type Answer struct {
	Name string
}

type Test struct {
	Name   string
	Number int
}

var Config struct {
	Answer    string   `toml:"answer"`
	Judge     string   `toml:"judge"`
	Verify    string   `toml:"verify"`
	Checker   string   `toml:"checker"`
	TimeLimit float64  `toml:"timeLimit"`
	Answers   []Answer `toml:"answers"`
	Tests     []Test   `toml:"tests"`
}

var (
	BaseFP      string
	MakerFP     string
	AnsFP       string
	BufFP       string
	ResFP       string
	InFP, OutFP string
	OtherFP     string
)

func init() {
	var fpath string
	flag.StringVar(&fpath, "toml", "", "toml file path")
	flag.Parse()

	log.SetLevel(log.DebugLevel)
	log.WithFields(log.Fields{"toml": fpath}).Info("Start")
	if _, err := toml.DecodeFile(fpath, &Config); err != nil {
		log.Fatal(err)
	}

	BaseFP, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	MakerFP = fp.Join(BaseFP, "maker")
	AnsFP = fp.Join(BaseFP, "answer")
	BufFP = fp.Join(BaseFP, "soj_"+rmExt(fpath))
	ResFP = fp.Join(BaseFP, "soj_"+rmExt(fpath)+"_case")
	InFP = fp.Join(ResFP, "in")
	OutFP = fp.Join(ResFP, "out")
	OtherFP = fp.Join(ResFP, "other")
}
