package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

var (
	target   = flag.String("target", "empty", "specify the site address")
	filetype = flag.String("type", "jpg", "specify the file type")
	arr      = [][]byte{}
	coll     = 0
)

func URLdomain(str string) (string, string) {
	url, err := url.Parse(str)
	if err != nil {
		fmt.Println(err)
	}
	return url.Hostname(), url.Scheme
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func filenameGen() string {
	const base64range string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	var id []byte

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 8; i++ {
		id = append(id, base64range[rand.Intn(37)])
	}

	return string(id)
}

func scrape() {
	collector := colly.NewCollector(
		colly.IgnoreRobotsTxt(),
		colly.Async(true),
		colly.MaxDepth(1),
	)

	hostname, _ := URLdomain(*target)
	collector.OnHTML("a", func(e *colly.HTMLElement) {
		src := e.Attr("href")

		if strings.Contains(src, "."+*filetype) {
			prefix := "http://"
			_, scheme := URLdomain(src)
			if scheme == "" {
				fmt.Println(prefix + hostname + src)
				collector.Visit(prefix + hostname + src)
			} else {
				fmt.Println(prefix + src)
				collector.Visit(prefix + src)
			}
			coll++
		}
	})

	collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode == 200 {
			data := r.Body
			arr = append(arr, data)
		}
	})

	collector.Visit(*target)
	collector.Wait()
}

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Usage of crawler:")
		flag.PrintDefaults()
	}
	flag.Parse()

	scrape()
	fmt.Println("Address:\t", *target)
	fmt.Println("File type:\t", *filetype)
	fmt.Println("Total links:\t\t", coll)

	for _, v := range arr {
		if coll != 0 {
			filename := filenameGen()
			if !Exists(filename) {
				err := ioutil.WriteFile(fmt.Sprintf("%s.%s", filename, *filetype), v, os.FileMode(0777))

				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	fmt.Println("Downloaded files:\t", len(arr)-1)
}
