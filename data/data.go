// Mix in data from different sources.
package data

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
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
	glog.V(2).Infof("Checking %s", p)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func path(name string) string {
	glog.V(2).Infof("Finding path to %s ...", name)
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
			glog.V(1).Infof("Found %s", p)
			return p
		}
		glog.V(1).Infof("%s not in %s", name, path)
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
		glog.V(1).Infof("Expanded %s to: %s", s, expanded)
	}
	return expanded
}

func ShortenName(s string) (string, bool) {
	glog.V(3).Infof("Shorten: %s", s)
	keyWords := strings.Split(expandAcronyms(s), " ")
	for i, kw := range keyWords {
		if _, exists := ExtraWords[strings.ToUpper(kw)]; exists {
			glog.V(1).Infof("Removing extra word in %s: %s", s, kw)
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
	glog.V(2).Infof("Merge: %s", r.Name)

	variations := []string{
		r.Name,
		strings.Join(strings.Split(shortName(expandAcronyms(r.Name)), " "), ""),
		shortName(r.Name),
		expandAcronyms(r.Name),
		shortName(expandAcronyms(r.Name)),
	}
	glog.V(2).Infof("Merge Variations: %v", strings.Join(variations, "|"))
	for _, name := range variations {
		mm := MMatches(name)
		glog.V(2).Infof("MMatches(%s) result: %v", name, mm)
		if len(mm) > 1 {
			// So, we have multiple matches. Perhaps the locale will help?
			glog.V(2).Infof("No unique for %s: %+v", name, mm)
			for _, m := range mm {
				// private knowledge
				if strings.Contains(r.ShortDesc, strings.Split(m, " - ")[1]) {
					glog.V(2).Infof("Lucky desc match: %s", m)
					r.M = M[m]
					return
				}
			}
		} else if len(mm) == 1 {
			glog.V(2).Infof("Match: %+v", mm)
			r.M = M[mm[0]]
			return
		}
	}
	glog.V(2).Infof("Unable to match: %+v", r)
}
