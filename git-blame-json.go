package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	reFirstLineOfHeader *regexp.Regexp
	reNonAlnum          *regexp.Regexp
)

func init() {
	reFirstLineOfHeader = regexp.MustCompile(`([a-fA-F0-9]{40}) (\d+) (\d+)( (\d+))?`)
	reNonAlnum = regexp.MustCompile(`[^a-zA-Z0-9]`)
}

func scanAsMap(reader io.Reader) ([]map[string]interface{}, error) {
	blames := []map[string]interface{}{}

	scanner := bufio.NewScanner(reader)

	lineNo := 0
	var b map[string]interface{} = nil
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		// First line of blamed line
		m := reFirstLineOfHeader.FindStringSubmatch(line)
		if m != nil {
			if b != nil {
				blames = append(blames, b)
			}
			b = map[string]interface{}{}
			b["sha1"] = m[1]
			b["original_line_number"] = m[2]
			b["final_line_number"] = m[3]
			continue
		}

		if b == nil {
			return nil, fmt.Errorf("syntax error on line %d", lineNo)
		}

		// content of actual line
		if strings.HasPrefix(line, "\t") {
			b["actual_line"] = line
			continue
		}

		// header
		items := strings.SplitN(line, " ", 2)
		if len(items) < 2 {
			return nil, fmt.Errorf("syntax error on line %d", lineNo)
		}

		key := items[0]
		value := items[1]
		key = reNonAlnum.ReplaceAllLiteralString(key, "_")
		b[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %s", err)
	}

	return blames, nil
}

func main() {
	blames, err := scanAsMap(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
		os.Exit(1)
	}

	err = json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"blame_lines": blames})
	if err != nil {
		fmt.Fprintln(os.Stderr, "json encoding error:", err)
		os.Exit(1)
	}
}
