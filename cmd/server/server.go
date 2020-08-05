// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"log"

	"github.com/tstromberg/campwiz/data"
	"github.com/tstromberg/campwiz/query"
	"github.com/tstromberg/campwiz/result"
)

type formValues struct {
	Dates    string
	Nights   int
	Distance int
	Standard bool
	Group    bool
	WalkIn   bool
	BoatIn   bool
}

type TemplateContext struct {
	Criteria query.Criteria
	Results  result.Results
	Form     formValues
}

func handler(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Incoming request: %+v", r)
	log.Println("Incoming request: %+v", r)
	var startDate = time.Now()
	var endDate time.Time
	var n, dist int
	var crit query.Criteria
	var results result.Results

	var err error
	if r.Method == "POST" {
		err = r.ParseForm()
		log.Println("RAW FORM IS %+v", r.Form)
		if err != nil {
			log.Println("error parsing the form: %v", err)
		}
		if r.Form["nights"] != nil {
			n, err = strconv.Atoi(r.Form["nights"][0])
			if err != nil {
				log.Println("Error parsing night %q : %v", r.Form["nights"][0], err)
			}
		}

		if r.Form["dates"] != nil {
			startDate, err = time.Parse("2006-01-02", r.Form["dates"][0])
			if err != nil {
				log.Println("Error parsing date %q : %v", r.Form["dates"][0], err)
			}

		}
		endDate = startDate.AddDate(0, 0, n)
		if r.Form["nights"] != nil {
			dist, err = strconv.Atoi(r.Form["distance"][0])
			if err != nil {
				log.Println("Error parsing distance %q : %v", r.Form["distance"][0], err)
			}
		}
		crit = query.Criteria{
			Dates:       []time.Time{startDate, endDate},
			Nights:      n,
			MaxDistance: dist,
			MaxPages:    10,
			// IncludeStandard: r.Form["standard"],
			// IncludeGroup:    r.Form["group"],
			// IncludeBoatIn:   r.Form["boat-in"],
			// IncludeWalkIn:   r.Form["walk-in"],
		}

		log.Println("Criteria is %+v", crit)

		results, err = query.Search(crit)
		log.Println("RESULTS: %+v", results)
		if err != nil {
			log.Println("Search error: %s", err)
		}

	} else {
		crit = query.Criteria{}

	}

	outTmpl, err := data.Read("http.tmpl")
	if err != nil {
		log.Println("Failed to read template: %v", err)
	}
	tmpl := template.Must(template.New("http").Parse(string(outTmpl)))
	ctx := TemplateContext{
		Criteria: crit,
		Results:  results,
		Form: formValues{
			Dates: startDate.Format("2006-01-02"),
		},
	}
	err = tmpl.ExecuteTemplate(w, "http", ctx)
	if err != nil {
		log.Println("template: %v", err)
	}
}

func init() {
	flag.Parse()
}
func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
