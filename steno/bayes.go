package main

import (
	"fmt"
	"github.com/jbrukh/bayesian"
	"regexp"
	"strings"
)

var tokSplitterPat = regexp.MustCompile(`[^\w]+`)

func tokenise(txt string) []string {
	out := []string{}
	for _, w := range tokSplitterPat.Split(txt, -1) {
		if len(w) > 2 {
			w = strings.ToLower(w)
			out = append(out, w)
		}
	}
	return out
}

func extractTags(arts ArtList) []bayesian.Class {
	found := map[string]struct{}{}

	for _, art := range arts {
		for _, tag := range art.Tags {
			found[tag] = struct{}{}
		}
	}
	tags := make([]bayesian.Class, 0, len(found))
	for tag, _ := range found {
		tags = append(tags, bayesian.Class(tag))
	}
	return tags
}

func Train(arts ArtList) (*bayesian.Classifier, error) {
	fmt.Printf("Train on %d arts\n", len(arts))
	allTags := extractTags(arts)
	fmt.Println(allTags)
	c := bayesian.NewClassifier(allTags...)

	for artCnt, art := range arts {
		txt := art.PlainTextContent()
		toks := tokenise(txt)
		for _, tag := range art.Tags {
			c.Learn(toks, bayesian.Class(tag))
		}

		//fmt.Printf("----------------------\n%s\n-----------------------", txt)
		//fmt.Println(toks)
		if artCnt%100 == 0 {
			fmt.Printf("trained %d/%d\n", artCnt, len(arts))
		}
	}

	err := c.WriteToFile("test.classifier")
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	} else {
		fmt.Printf("OK\n")
	}

	return nil, nil
}

func Classify(arts ArtList, store *Store) {
	c, err := bayesian.NewClassifierFromFile("test.classifier")

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Println(c.Classes)

	tagList := map[string]ArtList{}

	for artCnt, art := range arts {
		txt := art.PlainTextContent()
		toks := tokenise(txt)
		_, inx, _ := c.LogScores(toks)

		tag := string(c.Classes[inx])
		tagList[tag] = append(tagList[tag], art)

		if artCnt%100 == 0 {
			fmt.Printf("classified %d/%d\n", artCnt, len(arts))
		}
	}

	for tag, matching := range tagList {
		fmt.Printf("Apply %s to %d articles\n", tag, len(matching))
		_, err := store.AddTag(matching, tag)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}
	fmt.Printf("Done.\n")
}
