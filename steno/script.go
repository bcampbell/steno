package main

import (
	"bufio"
	//	"fmt"
	//	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type scriptLine struct {
	query string
	tags  []string
}

type script struct {
	Name  string
	Desc  string
	lines []scriptLine
}

var linePat = regexp.MustCompile(`^(?:([^#]+?)\s*=>\s*([^#]*?)\s*)?(?:#\s*(.*)\s*)?$`)

func strippedName(fullname string) string {
	b := path.Base(fullname)
	return b[0 : len(b)-len(path.Ext(b))]
}

func loadScript(filename string) (*script, error) {

	infile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer infile.Close()

	out := &script{Name: strippedName(filename)}
	lineNum := 0
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		bits := linePat.FindStringSubmatch(line)

		query, tagList, comment := bits[1], bits[2], bits[3]
		if lineNum == 1 && comment != "" {
			out.Desc = comment
		}

		query = strings.TrimSpace(query)
		if query != "" {
			tags := []string{}
			for _, tag := range strings.Split(tagList, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}

			out.lines = append(out.lines, scriptLine{query, tags})
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func loadScripts(dir string) (map[string]*script, error) {
	fileNames, err := filepath.Glob(path.Join(dir, "*.txt"))
	if err != nil {
		return nil, err
	}

	//	fmt.Printf("found %d scripts\n", len(fileNames))
	scripts := map[string]*script{}
	for _, fileName := range fileNames {
		s, err := loadScript(fileName)
		if err != nil {
			return nil, err
		}
		scripts[s.Name] = s
	}
	return scripts, nil
}
