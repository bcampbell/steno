package ft

import (
	"bufio"
	"fmt"
	"os/exec"
	"semprini/steno/steno/store"
	"strconv"
	"strings"
)

func Predict(db *store.Store, ftExe string, modelFilename string, threshold float64, progress func(float64)) (map[string]store.ArtList, error) {

	fmt.Printf("tag using %s\n", modelFilename)

	cmd := exec.Command(ftExe, "predict-prob", modelFilename, "-", "20")

	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(out)

	tags := map[string]store.ArtList{}

	totalArts := db.TotalArts()
	if totalArts == 0 {
		return nil, fmt.Errorf("No articles in db")
	}

	var scanErr error
	artit := db.IterateAllArts()
	// defer artit.Close()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	n := 0
	for artit.Next() {
		if progress != nil {
			perc := 100 * float64(n) / float64(totalArts)
			progress(perc)
		}
		art := artit.Cur()

		// pipe the article to fasttext
		dumpArt(art, in)
		// read a line of output
		if scanner.Scan() == false {
			// unexpected eof?
			break
		}
		var artTags []string
		artTags, scanErr = parseTags(scanner.Text(), threshold)
		if scanErr != nil {
			break
		}
		for _, tag := range artTags {
			tags[tag] = append(tags[tag], art.ID)
		}
		fmt.Printf("%d: %v\n", art.ID, artTags)
	}
	in.Close()

	cmdErr := cmd.Wait()

	if artit.Err() != nil {
		return nil, artit.Err()
	}

	if cmdErr != nil {
		return nil, cmdErr
	}
	if scanErr != nil {
		return nil, scanErr
	}

	return tags, nil
}

func applyTags(db *store.Store, tags map[string]store.ArtList, progress func(float64)) error {

	n := 0
	for tag, arts := range tags {
		if progress != nil {
			perc := 100 * float64(n) / float64(len(tags))
			progress(perc)
		}
		_, err := db.AddTags(arts, []string{tag})
		if err != nil {
			return err
		}
	}
	return nil
}

/*
 */

// __LABEL__<tag1> <prob1> __LABEL__<tag2> <prob2> ...
func parseTags(line string, threshold float64) ([]string, error) {
	labelPrefix := "__label__"
	tags := []string{}
	bits := strings.Fields(line)
	for i := 0; i < (len(bits) - 1); i += 2 {
		tag := strings.TrimPrefix(bits[i], labelPrefix)
		prob, err := strconv.ParseFloat(bits[i+1], 64)
		if err != nil {
			return nil, err
		}
		if prob >= threshold {
			tags = append(tags, tag)
		}
	}
	return tags, nil
}
