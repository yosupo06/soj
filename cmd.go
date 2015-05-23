package main

import (
	"bytes"
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"time"
)

func execCmd(s, fp string) ([]byte, []byte, time.Duration, error) {
	cmd := exec.Command("bash", "-c", s)
	if fp != "" {
		if _, err := os.Stat(fp); err != nil {
			log.Fatal(err)
		}
		cmd.Stdin, _ = os.Open(fp)
	}
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf
	c := make(chan error)
	go func() {
		c <- cmd.Run()
	}()
	select {
	case err := <-c:
		if err != nil {
			err = errors.New("RE")
		}
		return outBuf.Bytes(), errBuf.Bytes(), cmd.ProcessState.UserTime(), err
	case <-time.After(time.Duration(Config.TimeLimit) * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("Failed to kill:", err)
		}
		<-c
		return outBuf.Bytes(), errBuf.Bytes(), cmd.ProcessState.UserTime(), errors.New("TLE")
	}
	err := <-c
	//	err := cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), cmd.ProcessState.UserTime(), err
}

func fileCopy(dst, src string) {
	s, err := os.Open(src)
	defer s.Close()
	if err != nil {
		log.Fatal(err)
	}
	d, err := os.Create(dst)
	defer d.Close()
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(d, s)
	if err != nil {
		log.Fatal(err)
	}
}
