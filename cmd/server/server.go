// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/golang/glog"
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
	glog.Infof("Incoming request: %+v", r)
	var startDate = time.Now()
	var endDate time.Time
	var n, dist int
	var crit query.Criteria
	var results result.Results

	var err error
	if r.Method == "POST" {
		err = r.ParseForm()
		glog.Infof("RAW FORM IS %+v", r.Form)
		if err != nil {
			glog.Errorf("error parsing the form: %v", err)
		}
		if r.Form["nights"] != nil {
			n, err = strconv.Atoi(r.Form["nights"][0])
			if err != nil {
				glog.Errorf("Error parsing night %q : %v", r.Form["nights"][0], err)
			}
		}

		if r.Form["dates"] != nil {
			startDate, err = time.Parse("2006-01-02", r.Form["dates"][0])
			if err != nil {
				glog.Errorf("Error parsing date %q : %v", r.Form["dates"][0], err)
			}

		}
		endDate = startDate.AddDate(0, 0, n)
		if r.Form["nights"] != nil {
			dist, err = strconv.Atoi(r.Form["distance"][0])
			if err != nil {
				glog.Errorf("Error parsing distance %q : %v", r.Form["distance"][0], err)
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

		glog.Errorf("Criteria is %+v", crit)

		results, err = query.Search(crit)
		glog.V(1).Infof("RESULTS: %+v", results)
		if err != nil {
			glog.Errorf("Search error: %s", err)
		}

	} else {
		crit = query.Criteria{}

	}

	outTmpl, err := data.Read("http.tmpl")
	if err != nil {
		glog.Errorf("Failed to read template: %v", err)
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
		glog.Errorf("template: %v", err)
	}
}

func init() {
	flag.Parse()
}
func main() {
	http.HandleFunc("/", handler)
	glog.Fatal(http.ListenAndServe(":8080", nil))
}
