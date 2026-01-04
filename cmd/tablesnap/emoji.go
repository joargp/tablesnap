package main

import (
	"embed"
	"fmt"
	"image"
	"image/png"
	"strings"
	"unicode/utf8"
)

//go:embed twemoji/*.png
var twemojiFS embed.FS

// emojiMap maps emoji runes to their Twemoji filename (without .png)
var emojiMap = map[rune]string{
	'✅': "2705",
	'❌': "274c",
	'⭕': "2b55",
	'⚠': "26a0",
	'🔴': "1f534",
	'🟢': "1f7e2",
	'🟡': "1f7e1",
}

// loadTwemoji loads a Twemoji PNG by codepoint
func loadTwemoji(codepoint string) (image.Image, error) {
	data, err := twemojiFS.ReadFile("twemoji/" + codepoint + ".png")
	if err != nil {
		return nil, err
	}
	return png.Decode(strings.NewReader(string(data)))
}

// EmojiSegment represents either text or an emoji
type EmojiSegment struct {
	Text    string
	IsEmoji bool
	Emoji   rune
}

// ParseEmoji splits a string into text and emoji segments
func ParseEmoji(s string) []EmojiSegment {
	var segments []EmojiSegment
	var textBuf strings.Builder
	
	for _, r := range s {
		if _, ok := emojiMap[r]; ok {
			// Flush text buffer
			if textBuf.Len() > 0 {
				segments = append(segments, EmojiSegment{Text: textBuf.String()})
				textBuf.Reset()
			}
			segments = append(segments, EmojiSegment{IsEmoji: true, Emoji: r})
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

// GetEmojiImage returns the Twemoji image for a rune
func GetEmojiImage(r rune) (image.Image, error) {
	codepoint, ok := emojiMap[r]
	if !ok {
		return nil, fmt.Errorf("emoji not found: %c", r)
	}
	return loadTwemoji(codepoint)
}

// HasEmoji checks if a string contains any supported emoji
func HasEmoji(s string) bool {
	for _, r := range s {
		if _, ok := emojiMap[r]; ok {
			return true
		}
	}
	return false
}

// TextWidth returns approximate width without emoji
func TextWidthWithoutEmoji(s string) int {
	count := 0
	for _, r := range s {
		if _, ok := emojiMap[r]; !ok {
			count += utf8.RuneLen(r)
		}
	}
	return count
}
