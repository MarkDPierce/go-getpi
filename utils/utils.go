package utils

import (
	"log"
	"regexp"
	"strings"
)

var PathExtractor = regexp.MustCompile(`^(http[s]?:\/\/[^\s]+)(\/?[^?#]+)?`)

func DeterminePath(path []string) string {
	if len(path) > 0 && path[0] != "" {
		log.Print("Path is defined in config")
		return path[0]
	}
	log.Print("Path is /admin/")
	return "/admin/"
}

func ExtractIncludedPath(baseUrl, p string) (string, string) {
	includedPath := PathExtractor.FindStringSubmatch(baseUrl)
	if len(includedPath) > 0 {
		if len(includedPath) > 1 && len(includedPath[1]) > 0 && len(includedPath) > 2 && len(includedPath[2]) > 0 {
			baseUrl = includedPath[1]
			p = TrimTrailingSlash(includedPath[2]) + p
		}
	}
	return baseUrl, p
}

func CombineBaseURLAndPath(baseUrl, p string) string {
	return TrimTrailingSlash(baseUrl) + TrimTrailingSlash(p)
}

func TrimTrailingSlash(url string) string {
	if strings.HasSuffix(url, "/") {
		return url[:len(url)-1]
	}
	return url
}
