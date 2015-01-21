package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type scriptLine struct {
	query string
	op    string
	tags  []string
}

func (l *scriptLine) String() string {
	return l.query + "=>" + l.op + " " + strings.Join(l.tags, " ")
}

type script struct {
	Name  string
	Desc  string
	lines []scriptLine
}

// Apply script to a store
func (s *script) Run(store *Store) error {
	dbug.Printf("running script '%s'\n", s.Name)
	for _, line := range s.lines {
		matching, err := store.Search(line.query)
		if err != nil {
			return fmt.Errorf("Bad query (%s): %s", line.query, err)
		}

		dbug.Printf("%s (Matched %d)\n", line.String(), len(matching))

		switch line.op {
		case "tag":
			for _, tag := range line.tags {
				_, err := store.AddTag(matching, tag)
				if err != nil {
					return fmt.Errorf("tag error (during query '%s'): %s", line.query, err)
				}
			}
		case "untag":
			for _, tag := range line.tags {
				_, err := store.RemoveTag(matching, tag)
				if err != nil {
					return fmt.Errorf("untag error (during query '%s'): %s", line.query, err)
				}
			}
		case "delete":
			err := store.Delete(matching)
			if err != nil {
				return fmt.Errorf("error deleting (during query '%s'): %s", line.query, err)
			}
		}
	}
	return nil
}

var linePat = regexp.MustCompile(`(?i)^(?:([^#]+?)\s*=>\s*(tag|untag|delete)\s+([^#]*?)\s*)?(?:#\s*(.*)\s*)?$`)

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
		if bits == nil {
			return nil, fmt.Errorf("Syntax error in script '%s', line %d", filename, lineNum)
		}
		query, op, tagList, comment := bits[1], bits[2], bits[3], bits[4]
		if lineNum == 1 && comment != "" {
			// if there is a comment on the first line, use it as description
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

			op := strings.ToLower(op)
			out.lines = append(out.lines, scriptLine{query, op, tags})
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func loadScripts(dir string) ([]*script, error) {
	fileNames, err := filepath.Glob(path.Join(dir, "*.txt"))
	if err != nil {
		return nil, err
	}

	//	fmt.Printf("found %d scripts\n", len(fileNames))
	scripts := []*script{}
	for _, fileName := range fileNames {
		s, err := loadScript(fileName)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, s)
	}
	return scripts, nil
}
