package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"encoding/json"
	"regexp"
	"sort"
)

func main() {
	host := os.Getenv("HOST")
	u, e := url.Parse(host)
	if e != nil {
		panic("can't read HOST url")
	}

	var reader io.ReadCloser

	if u.Scheme == "http" {
		if e != nil {
			log.Fatalf("unable to parse host data url %s", host)
		}
		r, e2 := http.Get(u.String())

		if e2 != nil {
			log.Fatalf("unable to load url %s\n error: %#v\n", u.String(), e2)
		}

		log.Printf("building index from %s\n", u.String())
		reader = r.Body
	} else if u.Scheme == "file" {
		log.Println(fmt.Sprintf("Opening %s", u.Path))
		f, err := os.Open(u.Path)
		if err != nil {
			panic("can't read file")
		}
		reader = ioutil.NopCloser(bufio.NewReader(f))
	}

	log.Println("ingesting host 20mers")
	m := NewOfftargetMatcher(bufio.NewScanner(reader))
	reader.Close()

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		t := r.URL.Query().Get("targets")
		l := r.URL.Query().Get("limits")

		targets := strings.Split(t, ",")
		limits := strings.Split(l, ",")
		if len(targets) < 1 {
			http.Error(w, "please specify targets parameter as a comma separated list", 400)
			return
		}
		if len(limits) != 3 {
			http.Error(w, "please specify limits parmeter as a comma separated list of 3 items", 400)
			return
		}

		lim_c5, _ := strconv.Atoi(limits[0])
		lim_c10, _ := strconv.Atoi(limits[1])
		lim_c20, _ := strconv.Atoi(limits[2])

		for _, t := range targets {
			b := m.MatchForward(t, lim_c5, lim_c10, lim_c20)
			// logging here is too expensive
			//log.Printf("limits: %d %d %d targets %#v %t\n", lim_c5, lim_c10, lim_c20, t, b)
			fmt.Fprintf(w, "%s %t\n", t, b)
		}

	})

	http.HandleFunc("/api/2/search", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			http.Error(writer, "This endpoint only supports POST requests", 400)
			return
		}
		
		type SearchQuery struct {
			Queries []string
			Radius string
			Return_locations bool
		}

		log.Println("/api/2/search query")
		
		var query SearchQuery
		
		if request.Body == nil {
			http.Error(writer, "Please specify search parameters (search, radius, return_locations) in request body", 400)
			return
		}

		err := json.NewDecoder(request.Body).Decode(&query);
		if err != nil {
			log.Println(err.Error())
			http.Error(writer, "Error decoding search parameters. Expected JSON in request body with: queries, list of strings; radius, string specifying search radius, return_locations, bool\n" + err.Error(), 400)
			return
		}

		// We can only return locations if we were given an input file on disk
		if query.Return_locations && u.Scheme != "file" {
			http.Error(writer, "This server doesn't support return_locations", 400)
			return
		}
		
		
		// validate search query
		for _, protospacer := range query.Queries {
			if len(protospacer) != 20 {
				http.Error(writer, "queries parameter must be list of 20-mers", 400)
				return
			}
		}

		var radiiRe = regexp.MustCompile(`(\d+)_(\d+)_(\d+)`)
		radii := radiiRe.FindStringSubmatch(query.Radius)

		if len(radii) != 4 {
			http.Error(writer, fmt.Sprintf("radius parameter must be 5_9_x or 5_10_x, got %s", query.Radius), 400)
			return
		}

		lim_c5, _ := strconv.Atoi(radii[1])
		lim_c10, _ := strconv.Atoi(radii[2])
		lim_c20, _ := strconv.Atoi(radii[3])

		if lim_c5 != 5 || (lim_c10 != 9 && lim_c10 != 10) || (lim_c20 < lim_c10) || lim_c20 > 20 {
			http.Error(writer, fmt.Sprintf("radius parameter must be 5_9_x, with 9 <= x <= 20, or 5_10_x, with 10 <= x <= 20, but got %s", query.Radius), 400)
			return
		}

		log.Println(fmt.Sprintf("Running query on %d protospacers with radius %d_%d_%d", len(query.Queries), lim_c5, lim_c10, lim_c20))

		if !query.Return_locations {
			for _, t := range query.Queries {
				b := m.MatchForward(t, lim_c5, lim_c10, lim_c20)
				// logging here is too expensive
				//log.Printf("limits: %d %d %d targets %#v %t\n", lim_c5, lim_c10, lim_c20, t, b)
				fmt.Fprintf(writer, "%s %t\n", t, b)
			}
		} else {
			type Match struct {
				Protospacer string
				Protospacer_code twentymer 
				Matched bool
				Location []string
			}

			var matches []Match
			
			for _, t := range query.Queries {
				matches = append(matches, Match{t, encode(t), m.MatchForward(t, lim_c5, lim_c10, lim_c20), make([]string, 1, 10)})
				// logging here is too expensive
				//log.Printf("limits: %d %d %d targets %#v %t\n", lim_c5, lim_c10, lim_c20, t, b)
				//fmt.Fprintf(writer, "%s %t\n", t, b)
			}

			// get locations of matches from our input file

			// sort matches based on the twentymers
			sort.Slice(matches, func(i, j int) bool {
				return matches[i].Protospacer_code < matches[j].Protospacer_code
			})

			f, err := os.Open(u.Path)
			if err != nil {
				panic(fmt.Sprintf("can't read file %s", u.Path))
			}
			
			reader := bufio.NewReader(f)

			line, err := reader.ReadString('\n')
			fields := strings.Fields(line)			
			log.Println(fields)
			log.Println(err)
			j := 0
			
			for j < len(matches) && err == nil {
				if !matches[j].Matched {
					j = j + 1
					continue
				}
				
				log.Println(fmt.Sprintf("Searching for locations for match %s %d", matches[j].Protospacer, matches[j].Protospacer_code))
				for matches[j].Protospacer_code != encode(fields[0]) {
					// read the next line, until we find the match
					line, err = reader.ReadString('\n')
					fields = strings.Fields(line)

				}

				matches[j].Location = fields[1:]
				log.Println(fmt.Sprintf("Match %d found location %s", j, matches[j].Location))				
				j = j + 1				
			}

			f.Close()

			writer.Header().Set("Content-Type", "application/json")
			jsonData, err := json.Marshal(matches)

			writer.Write(jsonData)

			log.Println(matches)
		}
	})
	
	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
