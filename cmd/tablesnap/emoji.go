package main

import (
	"embed"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

//go:embed twemoji/*.png
var twemojiFS embed.FS

// cacheDir returns the emoji cache directory
func emojiCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "tablesnap", "twemoji")
}

// emojiToCodepoint converts an emoji rune to hex codepoint
func emojiToCodepoint(r rune) string {
	return fmt.Sprintf("%x", r)
}

// loadEmojiFromCache tries to load emoji from cache
func loadEmojiFromCache(codepoint string) (image.Image, error) {
	path := filepath.Join(emojiCacheDir(), codepoint+".png")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

// loadEmojiFromBundle tries to load emoji from embedded bundle
func loadEmojiFromBundle(codepoint string) (image.Image, error) {
	data, err := twemojiFS.ReadFile("twemoji/" + codepoint + ".png")
	if err != nil {
		return nil, err
	}
	return png.Decode(strings.NewReader(string(data)))
}

// GetEmojiImage returns the emoji image, checking cache first then bundle
func GetEmojiImage(r rune) (image.Image, bool) {
	codepoint := emojiToCodepoint(r)
	
	// Try cache first (full set)
	if img, err := loadEmojiFromCache(codepoint); err == nil {
		return img, true
	}
	
	// Try bundle (minimal set)
	if img, err := loadEmojiFromBundle(codepoint); err == nil {
		return img, true
	}
	
	return nil, false
}

// IsEmoji checks if a rune is likely an emoji (simplified check)
func IsEmoji(r rune) bool {
	// Common emoji ranges
	return (r >= 0x1F300 && r <= 0x1F9FF) || // Misc Symbols, Emoticons, etc
		(r >= 0x2600 && r <= 0x26FF) ||       // Misc Symbols
		(r >= 0x2700 && r <= 0x27BF) ||       // Dingbats
		(r >= 0x2300 && r <= 0x23FF) ||       // Misc Technical
		r == 0x2705 || r == 0x274C ||         // ✅ ❌
		r == 0x2B55 || r == 0x26A0            // ⭕ ⚠
}

// EmojiSegment represents either text or an emoji
type EmojiSegment struct {
	Text      string
	IsEmoji   bool
	Emoji     rune
	Supported bool
}

// ParseEmoji splits a string into text and emoji segments
func ParseEmoji(s string) []EmojiSegment {
	var segments []EmojiSegment
	var textBuf strings.Builder
	
	for _, r := range s {
		if IsEmoji(r) {
			// Flush text buffer
			if textBuf.Len() > 0 {
				segments = append(segments, EmojiSegment{Text: textBuf.String()})
				textBuf.Reset()
			}
			_, supported := GetEmojiImage(r)
			segments = append(segments, EmojiSegment{IsEmoji: true, Emoji: r, Supported: supported})
		} else {
			textBuf.WriteRune(r)
		}
	}
	
	// Flush remaining text
	if textBuf.Len() > 0 {
		segments = append(segments, EmojiSegment{Text: textBuf.String()})
	}
	
	return segments
}

// HasCachedEmojis checks if full emoji set is installed
func HasCachedEmojis() bool {
	info, err := os.Stat(emojiCacheDir())
	return err == nil && info.IsDir()
}
