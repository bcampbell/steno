package ft

import (
	"bufio"
	//	"bytes"
	"fmt"
	"github.com/bcampbell/steno/store"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

type TrainingParams struct {
	FasttextExe string
	Epoch       int
}

func BuildModel(db *store.Store, modelFilename string, params *TrainingParams, progress func(float64)) error {

	tmpfile, err := ioutil.TempFile("", "stenoft")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name()) // clean up

	// dump out tagged articles to tmpfile in fasttext format

	err = dumpTagged(db, tmpfile)
	if err != nil {
		return err
	}
	tmpfile.Close()

	// now run fasttext over them
	err = train(tmpfile.Name(), modelFilename, params, progress)
	if err != nil {
		return err
	}
	return nil
}

// invoke fasttext to build a model using a training set
func train(inFileName, outFileName string, params *TrainingParams, progress func(perc float64)) error {
	// eg ~/proj/fastText/fasttext supervised -input eu1.dump -output eu1.model -epoch 500 -wordNgrams 2
	args := []string{
		"supervised",
		"-input", inFileName,
		"-output", outFileName,
		"-epoch", strconv.Itoa(params.Epoch),
		"-wordNgrams", "2"}

	cmd := exec.Command(params.FasttextExe, args...)

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
