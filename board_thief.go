package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func URLdomain(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}

func main() {
	var target string
	flag.StringVar(&target, "target", "empty", "specify the site address")

	flag.Usage = func() {
		fmt.Fprintln(
			os.Stderr, "Usage: thief [--arg] [value]\n",
			"The program parses media files from network boards\n",
			"Arguments:",
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	if target == "empty" {
		flag.Usage()
		os.Exit(1)
	}

	resp, err := http.Get(target)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	mediaPattern := `src="([^"]+\.(jpg|png|gif|jpeg|mp4|mov|webm|svg))"`
	mediaRegex := regexp.MustCompile(mediaPattern)
	mediaMatches := mediaRegex.FindAllStringSubmatch(string(body), -1)

	quantity, downloaded := 0, 0
	for _, match := range mediaMatches {
		url := match[1]
		quantity++

		if !strings.HasPrefix(url, "http") {
			hostname, err := URLdomain(target)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			url = "https://" + hostname + url
		}

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		defer resp.Body.Close()

		filename := url[strings.LastIndex(url, "/")+1:]
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println("Error creating", filename, ":", err)
			continue
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println("Error writing", filename, ":", err)
			continue
		} else {
			downloaded++
		}
	}

	fmt.Printf("Total:\t%d\nDownloaded:\t%d\n", quantity, downloaded)
}
