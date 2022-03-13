package parse

import (
	"fmt"
	"regexp"
	"strconv"
)

/**
 * Parses input with the given regular expression and returns the
 * group values defined in the expression.
 */
func params(regEx *regexp.Regexp, input string) (paramsMap map[string]string) {
	match := regEx.FindStringSubmatch(input)

	paramsMap = make(map[string]string)
	for i, name := range regEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

// RENumbered regular expression for just a number
var RENumbered = regexp.MustCompile(`\{(?P<NUMBER>\d+)\}`)

// RENumberedWithNoExtension number for token indicating no file extension
var RENumberedWithNoExtension = regexp.MustCompile(`\{(?P<NUMBER>\d)+\.\}`)

// RENumberedBasename number for token indicating basename
var RENumberedBasename = regexp.MustCompile(`\{(?P<NUMBER>\d+)\/\}`)

// RENumberedDirname number for token indicating dirname
var RENumberedDirname = regexp.MustCompile(`\{(?P<NUMBER>\d+)\/\/\}`)

// RENumberedBasenameNoExtension number for token indicating basename with no extension
var RENumberedBasenameNoExtension = regexp.MustCompile(`\{(?P<NUMBER>\d+)\/\.\}`)

// RERange regular expression for a range such as {0..9}
var RERange = regexp.MustCompile(`\{(?P<START>\d+)\.\.(?P<END>\d+)\}`)

// NumberFromToken get a number from a token
func NumberFromToken(re *regexp.Regexp, input string) (found bool, number int, err error) {
	params := params(re, input)

	numberStr := params["NUMBER"]
	if params["NUMBER"] != "" {
		number, err = strconv.Atoi(numberStr)
		if err != nil {
			return
		}
		found = true
		return
	}
	return
}

// Range get a range from its token
func Range(input string) (rng []string, err error) {
	params := params(RERange, input)
	if params["START"] != "" && params["END"] != "" {
		var start, end int
		start, err = strconv.Atoi(params["START"])
		if err != nil {
			return
		}
		end, err = strconv.Atoi(params["END"])
		if err != nil {
			return
		}
		if start > end {
			err = fmt.Errorf("range %s has start %d > end %d", input, start, end)
			return
		}
		for i := start; i <= end; i++ {
			rng = append(rng, fmt.Sprint(i))
		}
	} else {
		err = fmt.Errorf("input %s start and/or end not found", input)
		return
	}

	return
}
