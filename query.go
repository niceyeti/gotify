/*
Returns: 0 on success, 1 on failure (no results)
Outputs: if found, output location and info



*/

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/pkg/errors"
)

func getHtml(ctx context.Context, url string, client *http.Client) (raw string, err error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		fmt.Printf("Error %s", err)
		return
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	raw = string(body)
	return
}

func parsePage(page string) (doc *html.Node, err error) {
	doc, err = html.Parse(strings.NewReader(page))
	if err != nil {
		log.Fatal(err)
	}
	return
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)

	err := html.Render(w, n)
	if err != nil {
		return ""
	}

	return buf.String()
}

func parsePullAndSave(page string) (r *result, err error) {
	var doc *html.Node
	doc, _ = parsePage(page)

	foundList := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" && n.FirstChild != nil {
			for _, attr := range n.Attr {
				if attr.Key == "id" && strings.Contains(attr.Val, "yard_locations_Year") {
					s := strings.Split(renderNode(n), "<span>")[1]
					s = strings.Split(s, "</span>")[0]
					if year, err := strconv.Atoi(s); err == nil {
						//fmt.Println("HIT", year)
						if year > 1984 && year <= 1988 {
							foundList = fmt.Sprintf("%s %d", foundList, year)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if foundList != "" {
		r = &result{
			description: foundList,
		}
	}

	return
}

func concat(inErr error, msg string) error {
	if inErr == nil {
		return fmt.Errorf(msg)
	}
	return fmt.Errorf("%w: %s", inErr, msg)
}

var ErrQueryFailure = errors.New("some queries failed")

func runQueries(ctx context.Context, sources []*source) (results []*result, err error) {
	queryTimeout := time.Duration(10) * time.Second
	cli := &http.Client{Timeout: queryTimeout}

	// Errors in the context of this loop are aggregated as a single error through their Error() string.
	for _, source := range sources {
		var page string
		var innerErr error
		requestCtx, _ := context.WithTimeout(ctx, queryTimeout)
		if page, innerErr = getHtml(requestCtx, source.url, cli); innerErr != nil {
			err = concat(err, innerErr.Error())
			continue
		}

		var r *result
		if r, innerErr = source.parseFn(page); innerErr != nil {
			err = concat(err, innerErr.Error())
			continue
		}

		if r != nil {
			r.source = source
			results = append(results, r)
		}
	}

	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), ErrQueryFailure)
	}

	return
}

type (
	result struct {
		description string
		source      *source
	}

	source struct {
		url      string
		entities string
		parseFn  func(string) (*result, error)
	}
)

/*
TODO:
	1) spalding parser: https://spaldings.hollanderstores.com/used-auto-parts/1987/nissan/maxima/brakes/536-caliper/536-58108l-left-rear,-
	2) Picknpull parser: https://www.picknpull.com/check-inventory/vehicle-search?make=234&model=4370&distance=25000&zip=99163&year=1984-1988
	3) craigslist
		https://pullman.craigslist.org/search/sss?query=nissan+maxima&searchNearby=2&nearbyArea=217&nearbyArea=52&nearbyArea=661&nearbyArea=322&nearbyArea=660&nearbyArea=659&nearbyArea=662&nearbyArea=324&nearbyArea=654&nearbyArea=656&nearbyArea=655&nearbyArea=9&nearbyArea=232&nearbyArea=2&nearbyArea=461&nearbyArea=95&nearbyArea=325&nearbyArea=246&max_auto_year=1988
		https://pullman.craigslist.org/search/sss?query=toyota+pickup&searchNearby=2&nearbyArea=217&nearbyArea=52&nearbyArea=661&nearbyArea=322&nearbyArea=660&nearbyArea=659&nearbyArea=662&nearbyArea=324&nearbyArea=654&nearbyArea=656&nearbyArea=655&nearbyArea=9&nearbyArea=232&nearbyArea=2&nearbyArea=461&nearbyArea=95&nearbyArea=325&nearbyArea=246&min_auto_year=1984&max_auto_year=1988
*/
var (
	sources = []*source{
		// Search for maximas
		{
			url:      "https://newautopart.net/includes/pullandsave/spokane/yard_locationslist.php?cmd=search&t=yard_locations&psearch=maxima&psearchtype=",
			entities: "maxima",
			parseFn:  parsePullAndSave,
		},
		// Search for maximas
		{
			url:      "https://newautopart.net/includes/pullandsave/mead/yard_locationslist.php?cmd=search&t=yard_locations&psearch=maxima&psearchtype=",
			entities: "maxima",
			parseFn:  parsePullAndSave,
		},
		// Search for toyotas
		{
			url:      "https://newautopart.net/includes/pullandsave/spokane/yard_locationslist.php?cmd=search&t=yard_locations&psearch=toyota&psearchtype=",
			entities: "toyota",
			parseFn:  parsePullAndSave,
		},
		// Search for toyotas
		{
			url:      "https://newautopart.net/includes/pullandsave/mead/yard_locationslist.php?cmd=search&t=yard_locations&psearch=toyota&psearchtype=",
			entities: "toyota",
			parseFn:  parsePullAndSave,
		},
	}
)

func main() {
	var err error
	if err = os.MkdirAll("results", 0o777); err != nil {
		fmt.Println(err.Error())
		return
	}

	// Future: use a proper logging interface instead.
	// Teeing to stdout like this is not thread safe.
	var errf *os.File
	if errf, err = os.Create("results/err.txt"); err != nil {
		fmt.Println(err.Error())
		return
	}
	defer errf.Close()
	// Echo errors to stdout
	stderr := io.MultiWriter(errf, os.Stdout)

	var logf *os.File
	if logf, err = os.Create("results/log.txt"); err != nil {
		fmt.Fprintf(stderr, "%s", err.Error())
		return
	}
	defer logf.Close()
	// Echo log to stdout
	log := io.MultiWriter(logf, os.Stdout)

	var results []*result
	if results, err = runQueries(context.Background(), sources); err != nil {
		fmt.Fprintf(stderr, "%s", err.Error())
	}

	fmt.Fprintf(stderr, "%d sources, %d hits", len(sources), len(results))
	for _, r := range results {
		if _, err = fmt.Fprintf(log, "\nSource entity '%s', found at %s:\n    %s\n", r.source.entities, r.source.url, r.description); err != nil {
			fmt.Fprintf(stderr, "%s", err.Error())
		}
	}

	retValue := 0
	if len(results) != len(sources) {
		retValue = 1
	}
	os.Exit(retValue)
}
