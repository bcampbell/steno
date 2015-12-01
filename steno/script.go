package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"semprini/steno/steno/store"
	"strings"
)

type scriptLine struct {
	srcLine int
	query   string
	op      string
	params  []string
}

func (l *scriptLine) String() string {
	return l.query + "=>" + l.op + " " + strings.Join(l.params, " ")
}

type script struct {
	Category string // taken from subdirectory
	Name     string
	Desc     string
	lines    []scriptLine
}

type ProgressFunc func(expected int, completed int, msg string)

// Apply script to a store
func (s *script) Run(store *store.Store, progress ProgressFunc) error {
	dbug.Printf("running script '%s'\n", s.Name)

	for lineNum, line := range s.lines {
		if progress != nil {
			progress(len(s.lines), lineNum, fmt.Sprintf("running %s: %d/%d", s.Name, lineNum+1, len(s.lines)))
		}
		matching, err := store.Search(line.query)
		if err != nil {
			return fmt.Errorf("Bad query on line %d (%s): %s", line.srcLine, line.query, err)
		}

		dbug.Printf("%s (Matched %d)\n", line.String(), len(matching))

		switch line.op {
		case "tag":
			tags := line.params
			_, err := store.AddTags(matching, tags)
			if err != nil {
				return fmt.Errorf("tag error (during query '%s'): %s", line.query, err)
			}
		case "untag":
			tags := line.params
			_, err := store.RemoveTags(matching, tags)
			if err != nil {
				return fmt.Errorf("untag error (during query '%s'): %s", line.query, err)
			}
		case "delete":
			err := store.Delete(matching, nil)
			if err != nil {
				return fmt.Errorf("error deleting (during query '%s'): %s", line.query, err)
			}
		}

	}
	return nil
}

var linePat = regexp.MustCompile(`(?i)^(?:([^#]+?)\s*=>\s*(tag|untag|delete)(?:\s+([^#]*?))?\s*)?(?:#\s*(.*)\s*)?$`)

func strippedName(fullname string) string {
	b := filepath.Base(fullname)

	return b[0 : len(b)-len(filepath.Ext(b))]
}

func loadScript(filename string) (*script, error) {

	infile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer infile.Close()

	out := &script{
		Category: filepath.Base(filepath.Dir(filename)),
		Name:     strippedName(filename),
	}
	lineNum := 0
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		bits := linePat.FindStringSubmatch(line)
		if bits == nil {
			return nil, fmt.Errorf("Syntax error in script '%s', line %d", filename, lineNum)
		}
		query, op, paramPart, comment := bits[1], bits[2], bits[3], bits[4]
		if lineNum == 1 && comment != "" {
			// if there is a comment on the first line, use it as description
			out.Desc = comment
		}

		query = strings.TrimSpace(query)
		if query != "" {
			params := strings.Fields(paramPart)
			op := strings.ToLower(op)
			l := scriptLine{srcLine: lineNum, query: query, op: op, params: params}
			//			fmt.Println(l, len(l.params))
			out.lines = append(out.lines, l)
			// check for dodgy cut&paste detritus
			if strings.ContainsAny(query, "“”") {
				dbug.Printf("WARNING %s (line %d): query has dodgy quotes: %s\n", filename, lineNum, query)
			}
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// load the simplified CSV-based script format
func loadCSVScript(filename string) (*script, error) {

	infile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer infile.Close()

	rdr := csv.NewReader(infile)
	// read header line

	header, err := rdr.Read()
	if err != nil {
		return nil, err
	}

	out := &script{
		Category: filepath.Base(filepath.Dir(filename)),
		Name:     strippedName(filename),
	}

	lineNum := 1
	for {

		row, err := rdr.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}

		}

		frags := []string{}
		tags := []string{}
		for i, col := range header {
			v := strings.TrimSpace(row[i])
			if col == "TAG" {
				tags = append(tags, strings.Fields(strings.ToLower(v))...)
			} else {
				if strings.ContainsRune(v, ' ') {
					v = `"` + v + `"`
				}
				frags = append(frags, col+":"+v)
			}
		}

		q := strings.Join(frags, " ")
		if q == "" {
			dbug.Printf("WARNING %s (line %d): empty query. Ignoring\n", filename, lineNum)
			continue
		}
		if len(tags) == 0 {
			dbug.Printf("WARNING %s (line %d): no tags. Ignoring\n", filename, lineNum)
			continue
		}
		l := scriptLine{srcLine: lineNum, query: q, op: "tag", params: tags}
		out.lines = append(out.lines, l)
	}

	//fmt.Printf("%+v\n", out)

	return out, nil
}

func loadScripts(dir string) ([]*script, error) {

	fileNames := []string{}
	err := filepath.Walk(dir, func(fileName string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(fileName)
		if ext != ".txt" && ext != ".csv" {
			dbug.Printf("WARNING ignoring %s\n", fileName)
			return nil
		}
		fileNames = append(fileNames, fileName)
		return nil
	})

	if err != nil {
		return nil, err
	}

	//	fmt.Printf("found %d scripts\n", len(fileNames))
	scripts := []*script{}
	for _, fileName := range fileNames {
		ext := filepath.Ext(fileName)
		var s *script
		var err error
		if ext == ".txt" {
			s, err = loadScript(fileName)
		} else if ext == ".csv" {
			s, err = loadCSVScript(fileName)
		} else {
			err = fmt.Errorf("Unknown script type %s\n", ext)
		}

		if err != nil {
			return nil, err
		}

		//	dbug.Printf("SCRIPT: [%s] %s\n", s.Category, s.Name)
		scripts = append(scripts, s)
	}
	return scripts, nil
}
