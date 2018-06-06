package cmd

import (
	"os"
	"regexp"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
)

func isTargetAliased(targetName string) bool {
	return strings.HasPrefix(targetName, "//")
}

var aliasPathMatcher *regexp.Regexp = regexp.MustCompile("//([A-Za-z]+)/(.+)")

func parse(aliasedName string) (alias string, blobName string, err error) {
	matches := aliasPathMatcher.FindStringSubmatch(aliasedName)
	jww.TRACE.Println("//alias/blobName matches ->", matches)
	if len(matches) != 3 {
		jww.ERROR.Println("Bad alias. Won't parse. ", aliasedName)
		jww.ERROR.Println("Matches:", matches)
		//TODO return an error instead of exiting here
		os.Exit(1)
	}
	return matches[1], matches[2], nil
}
