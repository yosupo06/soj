/*
TODO:ファイル名じゃなくてファイルそのものからハッシュを作成するように
*/
package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"hash/fnv"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type answer struct {
	Name string
}

type test struct {
	Name   string
	Number int
}

type tomlData struct {
	Answer    string   `toml:"answer"`
	Judge     string   `toml:"judge"`
	Verify    string   `toml:"verify"`
	TimeLimit float64  `toml:"timeLimit"`
	Answers   []answer `toml:"answers"`
	Tests     []test   `toml:"tests"`
}

var config struct {
	oo bool
	mz bool
	ve bool
}
var td tomlData

func init() {
	var fpath string
	flag.StringVar(&fpath, "toml", "", "toml file path")
	flag.BoolVar(&config.oo, "oo", false, "make output only")
	flag.BoolVar(&config.mz, "mz", false, "make zip file")
	flag.BoolVar(&config.ve, "verify", true, "verify")
	flag.Parse()
	log.SetLevel(log.DebugLevel)
	log.WithFields(log.Fields{"toml": fpath}).Info("Start")
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = toml.Decode(string(data), &td)
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll("soj_" + rmExt(fpath))
	os.Mkdir("soj_"+rmExt(fpath), 0755)
	err = os.Chdir("soj_" + rmExt(fpath))
	if err != nil {
		log.Fatal(err)
	}
}

var compileArg = map[string]string{
	".cpp": "clang++ -std=c++11 -O2 ../{{.Name}}.cpp -o {{.Name}}",
	".c":   "clang ../{{.Name}}.c -o {{.Name}}",
	".d":   "dmd -O ../{{.Name}}.d",
	".txt": ":",
}

var testArg = map[string]string{
	".cpp": "./{{.Name}} --seed={{.Seed}}",
	".d":   "./{{.Name}} --seed={{.Seed}}",
	".txt": "cat ../{{.Name}}/{{.Name}}{{.Seed}}.txt",
}

var execArg = map[string]string{
	".cpp": "./{{.Name}}",
	".c":   "./{{.Name}}",
	".d":   "./{{.Name}}",
	".txt": "cat ../{{.Name}}/{{.Name}}{{.Seed}}.txt",
}

func rmExt(s string) string {
	return strings.TrimSuffix(s, fp.Ext(s))
}

func caseName(s string, c int) string {
	return rmExt(s) + "_" + strconv.Itoa(c) + ".txt"
}

var sourcehash = make(map[string]uint32)

func compile(s string) {
	ext := fp.Ext(s)
	fnvhs := fnv.New32a()
	fnvhs.Write([]byte(s))
	sourcehash[s] = fnvhs.Sum32()
	if ext == ".txt" {
		sourcehash[s] = 0
	}
	tpl, _ := template.New("cmd").Parse(compileArg[ext])
	var b bytes.Buffer
	tpl.Execute(&b, map[string]string{
		"Name": rmExt(s)})
	_, errB, _, err := execCmd(b.String(), "")
	if err != nil {
		log.Error(string(errB))
		log.Fatal(err)
	}
}

func makeTest(s string, cs int) ([]byte, error) {
	ext := fp.Ext(s)
	tpl, _ := template.New("cmd").Parse(testArg[ext])
	var b bytes.Buffer
	cn := uint32(cs) + sourcehash[s]
	tpl.Execute(&b, map[string]string{
		"Name": rmExt(s), "Seed": strconv.FormatUint(uint64(cn), 10)})
	out, _, _, err := execCmd(b.String(), "")
	return out, err
}

func makeOutput(s, ans string) ([]byte, time.Duration, error) {
	ext := fp.Ext(s)
	tpl, _ := template.New("cmd").Parse(execArg[ext])
	var b bytes.Buffer
	tpl.Execute(&b, map[string]string{
		"Name": rmExt(s)})
	outB, _, du, err := execCmd(b.String(), "cases/in/"+ans)
	return outB, du, err
}

func verifyAnswer(s, ins string, c int) (string, time.Duration) {
	fn := caseName(ins, c)
	b, du, err := makeOutput(s, fn)
	if err != nil {
		return err.Error(), du
	}
	err = ioutil.WriteFile("answer/"+rmExt(s)+"/"+fn, b, 0644)
	if err != nil {
		log.Fatal(err)
	}
	cor, err := ioutil.ReadFile("cases/out/" + fn)
	if err != nil {
		log.Fatal(err)
	}
	if textDiff(string(b), string(cor)) {
		return "AC", du
	} else {
		return "WA", du
	}
}

func textDiff(inf, ouf string) bool {
	inl := strings.Fields(inf)
	oul := strings.Fields(ouf)
	if len(inl) != len(oul) {
		return false
	}
	for i, v := range oul {
		if inl[i] != v {
			return false
		}
	}
	return true
}

func makeCase() {
	log.Info("Start Maker Compile")
	for _, v := range td.Tests {
		compile(v.Name)
		log.WithFields(log.Fields{"Name": v.Name}).Info("Compiling ...")
	}
	log.Info("Start Make Input")
	os.RemoveAll("cases")
	os.MkdirAll("cases/in", 0755)
	if config.mz {
		os.MkdirAll("cases/inzip", 0755)
	}
	for _, v := range td.Tests {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		n := rmExt(v.Name)
		log.WithFields(log.Fields{
			"Name": n, "Number": v.Number}).Info("Make Input")
		for i := 0; i < v.Number; i++ {
			out, err := makeTest(v.Name, i)
			if err != nil {
				log.Error(err)
			}
			err = ioutil.WriteFile("cases/in/"+caseName(v.Name, i), out, 0644)
			if err != nil {
				log.Error(err)
			}
			if config.mz {
				f, err := w.Create(caseName(v.Name, i))
				if err != nil {
					log.Error(err)
				}
				_, err = f.Write(out)
				if err != nil {
					log.Error(err)
				}
			}
		}
		err := w.Close()
		if err != nil {
			log.Fatal(err)
		}
		if config.mz {
			err = ioutil.WriteFile("cases/inzip/"+n+".zip", buf.Bytes(), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
func verify() {
	if td.Verify == "" {
		log.Fatal("Verify is Empty")
	}
	log.Info("Start Verify")
	compile(td.Verify)
	log.WithFields(log.Fields{"Name": td.Verify}).Info("Compiling ...")
	for _, v := range td.Tests {
		n := rmExt(v.Name)
		log.WithFields(log.Fields{
			"Name": n, "Number": v.Number}).Info("Verify Input")
		for i := 0; i < v.Number; i++ {
			outB, _, err := makeOutput(td.Verify, caseName(v.Name, i))
			if len(outB) != 0 {
				log.Info(string(outB))
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func makeAnswer() {
	log.Info("Make Answer")
	log.WithFields(log.Fields{"Name": td.Answer}).Info("Compiling ...")
	compile(td.Answer)
	os.MkdirAll("cases/out", 0755)
	for _, v := range td.Tests {
		log.WithFields(log.Fields{
			"Name": v.Name, "Number": v.Number}).Debug("Make Answer")
		for i := 0; i < v.Number; i++ {
			outB, du, err := makeOutput(td.Answer, caseName(v.Name, i))
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile("cases/out/"+caseName(v.Name, i), outB, 0644)
			if err != nil {
				log.Fatal(err)
			}
			log.WithFields(log.Fields{
				"Time": int64(du / time.Millisecond)}).Debugf("seed = %d", i)
		}
	}
}

func main() {
	makeCase()

	if config.ve {
		verify()
	}

	if config.oo {
		return
	}

	makeAnswer()

	for _, v := range td.Answers {
		compile(v.Name)
		log.WithFields(log.Fields{"Name": v.Name}).Info("Compiling ...")
	}

	os.RemoveAll("answer")
	os.MkdirAll("answer", 0755)
	for _, a := range td.Answers {
		os.Mkdir("answer/"+rmExt(a.Name), 0755)
		log.WithFields(log.Fields{
			"Name": rmExt(a.Name)}).Info("Start Check")
		for _, t := range td.Tests {
			log.WithFields(log.Fields{
				"DataSet": t.Name}).Info("Checking...")
		L:
			for i := 0; i < t.Number; i++ {
				switch res, du := verifyAnswer(a.Name, t.Name, i); res {
				case "AC":
					log.WithFields(log.Fields{
						"Time": int64(du / time.Millisecond)}).Debugf("AC. seed = %d", i)
				case "WA", "RE", "TLE":
					log.WithFields(log.Fields{
						"Name": t.Name, "Number": i}).Warn(res)
					break L
				}
			}
		}
	}
	log.Info("End Check")
}
