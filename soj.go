/*
TODO:ファイル名じゃなくてファイルそのものからハッシュを作成するように
*/
package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-colorable"
	"hash/fnv"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var tp = make(map[string]*template.Template)

func init() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(colorable.NewColorableStdout())
	for k, c := range compileArg {
		t, err := template.New("comp").Parse(c)
		if err != nil {
			panic(err)
		}
		tp["compile"+k] = t
	}
	for k, c := range testArg {
		t, err := template.New("test").Parse(c)
		if err != nil {
			panic(err)
		}
		tp["test"+k] = t
	}
	for k, c := range execArg {
		t, err := template.New("exec").Parse(c)
		if err != nil {
			panic(err)
		}
		tp["exec"+k] = t
	}
	for k, c := range checkerArg {
		t, err := template.New("checker").Parse(c)
		if err != nil {
			panic(err)
		}
		tp["checker"+k] = t
	}
}

func rmExt(s string) string {
	return strings.TrimSuffix(s, fp.Ext(s))
}

func caseName(s string, c int) string {
	return rmExt(s) + "_" + strconv.Itoa(c) + ".txt"
}

var sourcehash = make(map[string]uint32)

func Command(ty, s, f string, flag map[string]string) (
	[]byte, []byte, time.Duration, error) {
	ext := fp.Ext(s)
	var b bytes.Buffer
	if flag == nil {
		flag = make(map[string]string)
	}
	flag["Name"] = rmExt(s)
	tp[ty+ext].Execute(&b, flag)
	return execCmd(b.String(), f)
}

func compile(s string) {
	fnvhs := fnv.New32a()
	fnvhs.Write([]byte(s))
	sourcehash[s] = fnvhs.Sum32()
	_, errB, _, err := Command("compile", s, "", nil)
	if err != nil {
		log.Error(string(errB))
		log.Fatal(err)
	}
}

func makeCase() {
	log.Info("Start Maker Compile")
	for _, v := range Config.Tests {
		if fp.Ext(v.Name) == ".txt" {
			continue
		}
		fileCopy(fp.Join(BufFP, v.Name), fp.Join(MakerFP, v.Name))
		log.WithField("Name", v.Name).Info("Start Compile")
		compile(v.Name)
	}
	log.Info("Start Make Input")
	os.MkdirAll(InFP, 0755)
	for _, v := range Config.Tests {
		n := rmExt(v.Name)
		log.WithFields(log.Fields{
			"Name": n, "Number": v.Number}).Info("Make Input")
		for i := 0; i < v.Number; i++ {
			if fp.Ext(v.Name) == ".txt" {
				fileCopy(fp.Join(InFP, caseName(v.Name, i)), fp.Join(MakerFP, n, caseName(v.Name, i)))
				continue
			}
			out, _, _, err := Command("test", v.Name, "",
				map[string]string{"Seed": strconv.Itoa(i),
					"Hash": strconv.FormatUint(uint64(sourcehash[v.Name]), 10)})
			if err != nil {
				log.Error(err)
			}
			err = ioutil.WriteFile(fp.Join(InFP, caseName(v.Name, i)), out, 0644)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func verify() {
	log.Info("Start Verify")
	fileCopy(fp.Join(BufFP, Config.Verify), fp.Join(MakerFP, Config.Verify))
	compile(Config.Verify)
	log.WithField("Name", Config.Verify).Info("Compiling ...")
	for _, v := range Config.Tests {
		n := rmExt(v.Name)
		log.WithFields(log.Fields{
			"Name": n, "Number": v.Number}).Info("Verify Input")
		for i := 0; i < v.Number; i++ {
			outB, _, _, err := Command("exec", Config.Verify, fp.Join(InFP, caseName(v.Name, i)), nil)
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
	log.WithField("Name", Config.Answer).Info("Compiling ...")
	if Config.Checker != "" {
		fileCopy(fp.Join(BufFP, Config.Checker), fp.Join(MakerFP, Config.Checker))
		compile(Config.Checker)
	}
	fileCopy(fp.Join(BufFP, Config.Answer), fp.Join(AnsFP, Config.Answer))
	compile(Config.Answer)
	os.MkdirAll(OutFP, 0755)
	for _, v := range Config.Tests {
		log.WithFields(log.Fields{
			"Name": v.Name, "Number": v.Number}).Debug("Make Answer")
		for i := 0; i < v.Number; i++ {
			cn := caseName(v.Name, i)
			outB, _, du, err := Command("exec", Config.Answer, fp.Join(InFP, cn), nil)
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile(fp.Join(OutFP, cn), outB, 0644)
			if err != nil {
				log.Fatal(err)
			}
			log.WithField("Time", int64(du/time.Millisecond)).Debugf("seed = %d", i)
			if ok, _ := check(fp.Join(InFP, cn), fp.Join(OutFP, cn), ""); ok {
				log.Info("OK")
			} else {
				log.Fatal("ERROR")
			}
		}
	}
}

func check(in, out, ans string) (bool, string) {
	if Config.Checker == "" {
		if ans == "" {
			return true, "AC"
		}
		o, err := ioutil.ReadFile(out)
		if err != nil {
			log.Fatal(err)
		}
		a, err := ioutil.ReadFile(ans)
		if err != nil {
			log.Fatal(err)
		}
		if textDiff(string(o), string(a)) {
			return true, "AC"
		} else {
			return false, "WA"
		}
	} else {
		_, e, _, err := Command("checker", Config.Checker, "",
			map[string]string{
				"Input":  in,
				"Output": out,
				"Answer": ans,
			})
		fmt.Print(string(e))
		if err != nil {
			return false, err.Error()
		}
		return true, "AC"
	}
}

func textDiff(out, ans string) bool {
	oul := strings.Fields(out)
	anl := strings.Fields(ans)
	if len(oul) != len(anl) {
		return false
	}
	for i, v := range oul {
		if anl[i] != v {
			return false
		}
	}
	return true
}

func checkOther() {
	for _, v := range Config.Answers {
		fileCopy(fp.Join(BufFP, v.Name), fp.Join(AnsFP, v.Name))
		compile(v.Name)
		log.WithField("Name", v.Name).Info("Compiling ...")
	}

	os.MkdirAll(OtherFP, 0755)
	for _, a := range Config.Answers {
		path := fp.Join(OtherFP, rmExt(a.Name))
		os.MkdirAll(path, 0755)
		log.WithField("Name", rmExt(a.Name)).Info("Start Check")
		for _, t := range Config.Tests {
			log.WithField("DataSet", t.Name).Info("Checking...")

			for i := 0; i < t.Number; i++ {
				cn := caseName(t.Name, i)
				outB, _, du, err := Command("exec", a.Name, fp.Join(InFP, cn), nil)
				if err != nil {
					log.WithFields(log.Fields{
						"Name": t.Name, "Number": i}).Warn(err)
					break
				}
				err = ioutil.WriteFile(fp.Join(path, cn), outB, 0644)
				if err != nil {
					log.Fatal(err)
				}
				if ok, mes := check(fp.Join(InFP, cn), fp.Join(path, cn), fp.Join(OutFP, cn)); ok {
					log.WithField("Time", int64(du/time.Millisecond)).Infof("AC seed=%d", i)
				} else {
					log.WithFields(log.Fields{
						"Name": t.Name, "Number": i}).Warn(mes)
					break
				}
			}
		}
	}
	log.Info("End Check")
}

func main() {
	if err := os.RemoveAll(BufFP); err != nil {
		log.Fatal(err)
	}
	if err := os.RemoveAll(ResFP); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(BufFP, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(ResFP, 0755); err != nil {
		log.Fatal(err)
	}
	os.Chdir(BufFP)
	makeCase()

	if Config.Verify == "" {
		log.Info("Skip Verify")
	} else {
		verify()
	}

	makeAnswer()

	checkOther()

}
