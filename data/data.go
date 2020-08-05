// Mix in data from different sources.
package data

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tstromberg/campwiz/result"
)

var (
	M map[string]result.MEntry

	Acronyms = map[string]string{
		"MT.": "MOUNT",
		"SB":  "STATE BEACH",
		"SRA": "STATE RECREATION AREA",
		"SP":  "STATE PARK",
		"CP":  "COUNTY PARK",
		"NP":  "NATIONAL PARK",
	}

	ExtraWords = map[string]bool{
		"&":          true,
		"(CA)":       true,
		"AND":        true,
		"AREA":       true,
		"CAMP":       true,
		"CAMPGROUND": true,
		"COUNTY":     true,
		"DAY":        true,
		"FOREST":     true,
		"FS":         true,
		"MONUMENT":   true,
		"NATIONAL":   true,
		"NATL":       true,
		"PARK":       true,
		"RECREATION": true,
		"REGIONAL":   true,
		"STATE":      true,
		"USE":        true,
	}
)

func exists(p string) bool {
	log.Println("Checking %s", p)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func path(name string) string {
	log.Println("Finding path to %s ...", name)
	binpath, err := os.Executable()
	if err != nil {
		binpath = "."
	}

	for _, d := range []string{
		"./",
		"../",
		"../../",
		filepath.Join(filepath.Dir(binpath)),
		filepath.Join(build.Default.GOPATH, "github.com/tstromberg/campwiz"),
	} {
		p := filepath.Join(d, "data", name)
		if exists(p) {
			log.Println("Found %s", p)
			return p
		}
		log.Println("%s not in %s", name, path)
	}
	return ""
}

// Find path to data, return data from it.
func Read(name string) ([]byte, error) {
	p := path(name)
	if p == "" {
		return nil, fmt.Errorf("Could not find %s", name)
	}
	return ioutil.ReadFile(p)
}

func expandAcronyms(s string) string {
	var words []string
	for _, w := range strings.Split(s, " ") {
		if val, exists := Acronyms[strings.ToUpper(w)]; exists {
			words = append(words, val)
		} else {
			words = append(words, w)
		}
	}
	expanded := strings.Join(words, " ")
	if expanded != s {
		log.Println("Expanded %s to: %s", s, expanded)
	}
	return expanded
}

func ShortenName(s string) (string, bool) {
	log.Println("Shorten: %s", s)
	keyWords := strings.Split(expandAcronyms(s), " ")
	for i, kw := range keyWords {
		if _, exists := ExtraWords[strings.ToUpper(kw)]; exists {
			log.Println("Removing extra word in %s: %s", s, kw)
			keyWords = append(keyWords[:i], keyWords[i+1:]...)
			return strings.Join(keyWords, " "), true
		}
	}
	return s, false
}

func shortName(s string) string {
	var shortened bool
	for {
		s, shortened = ShortenName(s)
		if !shortened {
			break
		}
	}
	return s
}

func Merge(r *result.Result) {
	log.Println("Merge: %s", r.Name)

	variations := []string{
		r.Name,
		strings.Join(strings.Split(shortName(expandAcronyms(r.Name)), " "), ""),
		shortName(r.Name),
		expandAcronyms(r.Name),
		shortName(expandAcronyms(r.Name)),
	}
	log.Println("Merge Variations: %v", strings.Join(variations, "|"))
	for _, name := range variations {
		mm := MMatches(name)
		log.Println("MMatches(%s) result: %v", name, mm)
		if len(mm) > 1 {
			// So, we have multiple matches. Perhaps the locale will help?
			log.Println("No unique for %s: %+v", name, mm)
			for _, m := range mm {
				// private knowledge
				if strings.Contains(r.ShortDesc, strings.Split(m, " - ")[1]) {
					log.Println("Lucky desc match: %s", m)
					r.M = M[m]
					return
				}
			}
		} else if len(mm) == 1 {
			log.Println("Match: %+v", mm)
			r.M = M[mm[0]]
			return
		}
	}
	log.Println("Unable to match: %+v", r)
}
