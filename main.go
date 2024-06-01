package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Track struct {
	Track       int    `json:"track"`
	Name        string `json:"name"`
	ChapterLink string `json:"chapter_link_dropbox"`
	Duration    string `json:"duration"`
	ChapterID   string `json:"chapter_id"`
	PostID      string `json:"post_id"`
	URL         string `json:"url"`
}

func main() {
	book := os.Getenv("FOLDER_NAME")
	if book == "" {
		book = "book"
	}
	trackFile := filepath.Join(book, "tracks.json")
	outputFolder := filepath.Join(book, "chapters")
	host := os.Getenv("HOST")
	if host == "" {
		host = "https://files02.tokybook.com/audio/"
	}

	if err := downloadBook(trackFile, outputFolder, host); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func downloadBook(trackFile, outputFolder, host string) error {
	content, err := os.ReadFile(trackFile)
	if err != nil {
		return err
	}

	tracks := []Track{}
	err = json.Unmarshal(content, &tracks)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d tracks\n", len(tracks))

	err = os.MkdirAll(outputFolder, os.ModePerm)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for _, track := range tracks {
		url := host + track.ChapterLink

		name := strings.ReplaceAll(track.Name, " ", "-")
		output := outputFolder + "/" + name
		go downloadChapter(url, output, &wg)
		wg.Add(1)
	}

	wg.Wait()
	return nil
}

func downloadChapter(url, output string, wg *sync.WaitGroup) {
	err := func() error {
		fmt.Printf("Downloading %s...\n", output)

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		file, err := os.Create(output)
		if err != nil {
			return err
		}
		defer file.Close()

		bytesWritten, err := io.Copy(file, resp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("Wrote %f MB\n", float64(bytesWritten)/(1<<20))
		return nil
	}()
	if err != nil {
		fmt.Println("Error during download: ", err)
	} else {
		fmt.Printf("Downloaded %s\n", output)
	}
	wg.Done()
}
