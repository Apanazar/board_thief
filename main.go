package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

func main() {
	targetURL := flag.String("target", "", "Target URL")
	flag.Parse()

	if *targetURL == "" {
		fmt.Println("Please provide a target URL using --target flag")
		return
	}

	body, err := fetchURLContent(*targetURL)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	matches := findMediaFiles(body)

	var wg sync.WaitGroup
	for _, match := range matches {
		filename := match[1]
		fileURL := generateFileURL(*targetURL, filename)

		wg.Add(1)
		go func(url, filepath string) {
			defer wg.Done()
			if err := downloadFile(url, filepath); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Successfully:", filepath)
			}
		}(fileURL, filename)
	}
	wg.Wait()
}

func fetchURLContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func findMediaFiles(content []byte) [][]string {
	re := regexp.MustCompile(`src/(\d+\.(?:jpg|png|gif|jpeg|mp4|mov|webm|svg))`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	return matches
}

func generateFileURL(baseURL, filename string) string {
	return strings.TrimSuffix(baseURL, "/1.html") + "/src/" + filename
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %d %s\n", resp.StatusCode, url)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
