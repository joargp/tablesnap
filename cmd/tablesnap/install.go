package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const twemojiURL = "https://github.com/twitter/twemoji/archive/refs/heads/master.zip"

func installEmojis() error {
	cacheDir := emojiCacheDir()
	
	// Check if already installed
	if HasCachedEmojis() {
		fmt.Println("Emoji pack already installed at", cacheDir)
		return nil
	}
	
	fmt.Println("Downloading Twemoji pack...")
	
	// Download zip
	resp, err := http.Get(twemojiURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Save to temp file
	tmpFile, err := os.CreateTemp("", "twemoji-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}
	tmpFile.Close()
	
	fmt.Println("Extracting 72x72 PNGs...")
	
	// Create cache dir
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	
	// Extract 72x72 PNGs
	r, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return err
	}
	defer r.Close()
	
	count := 0
	for _, f := range r.File {
		if strings.Contains(f.Name, "assets/72x72/") && strings.HasSuffix(f.Name, ".png") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			
			filename := filepath.Base(f.Name)
			outPath := filepath.Join(cacheDir, filename)
			
			outFile, err := os.Create(outPath)
			if err != nil {
				rc.Close()
				continue
			}
			
			io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			count++
		}
	}
	
	fmt.Printf("✅ Installed %d emoji to %s\n", count, cacheDir)
	return nil
}
