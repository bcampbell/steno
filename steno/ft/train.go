package main

import (
	"bufio"
	//	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

// invoke fasttext to build a model using a training set
func trainem(ftExe, inFileName, outFileName string, progress func(perc float64)) error {
	// eg ~/proj/fastText/fasttext supervised -input eu1.dump -output eu1.model -epoch 500 -wordNgrams 2
	args := []string{
		"supervised",
		"-input", inFileName,
		"-output", outFileName,
		"-epoch", "20", //"500",
		"-wordNgrams", "2"}

	cmd := exec.Command(ftExe, args...)

	// use a custom scanner to scan the output text for progress percentage
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	progPat := regexp.MustCompile("Progress: ([.0-9]+)%")
	splitter := func(data []byte, atEOF bool) (int, []byte, error) {
		loc := progPat.FindSubmatchIndex(data)
		if loc == nil {
			return 0, nil, nil
		}
		return loc[1], data[loc[2]:loc[3]], nil
	}
	scanner.Split(splitter)

	// go!
	err = cmd.Start()
	if err != nil {
		return err
	}
	for scanner.Scan() {
		if s, err := strconv.ParseFloat(scanner.Text(), 64); err == nil {
			if progress != nil {
				progress(s)
			}
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading fasttext output:", err)
	}

	//
	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
