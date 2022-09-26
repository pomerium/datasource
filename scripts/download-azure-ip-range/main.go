package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const confirmURL = "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	wd, err := getWorkingDirectory()
	if err != nil {
		return err
	}

	downloadURL, err := getDownloadURL(ctx)
	if err != nil {
		return err
	}

	err = saveAzureIPRanges(ctx, wd, downloadURL)
	if err != nil {
		return err
	}

	return nil
}

func getWorkingDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for wd != "/" {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		wd = filepath.Dir(wd)
	}

	return "", fmt.Errorf("go.mod not found")
}

func getDownloadURL(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, confirmURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating confirm URL request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching confirm URL: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return "", fmt.Errorf("invalid confirm URL result: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading confirm URL result: %w", err)
	}

	re := regexp.MustCompile(`failoverLink.*href="([^"]*)"`)
	match := re.FindStringSubmatch(string(body))
	if len(match) <= 1 {
		return "", fmt.Errorf("failed to find failover link in body")
	}

	return match[1], nil
}

func saveAzureIPRanges(ctx context.Context, wd string, downloadURL string) error {
	dst, err := os.Create(filepath.Join(wd, "internal", "wellknownips", "files", "azure.json.gz"))
	if err != nil {
		return fmt.Errorf("error creating azure.json.gz: %w", err)
	}
	defer dst.Close()

	gzw, err := gzip.NewWriterLevel(dst, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("error creating gzip writer: %w", err)
	}
	defer gzw.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return fmt.Errorf("invalid response status code from azure ip range url: %s", res.Status)
	}

	_, err = io.Copy(gzw, res.Body)
	return err
}
