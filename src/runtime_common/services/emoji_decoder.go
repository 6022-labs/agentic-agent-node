package services

import (
	"sort"
	"strings"

	"github.com/kyokomi/emoji/v2"
)

// EmojiDecoder maps emoji glyphs to shortcodes so messages round-trip through
// Slack/broker.
type EmojiDecoder struct {
	replacer *strings.Replacer
}

func NewEmojiDecoder() *EmojiDecoder {
	decoder := &EmojiDecoder{}
	decoder.replacer = decoder.buildDemojizeReplacer()
	return decoder
}

// buildDemojizeReplacer builds a Replacer from kyokomi's reverse codemap.
// Sort by glyph length desc so ZWJ-joined sequences beat their sub-glyphs.
func (d *EmojiDecoder) buildDemojizeReplacer() *strings.Replacer {
	revMap := emoji.RevCodeMap()

	type pair struct {
		unicode   string
		shortcode string
	}
	pairs := make([]pair, 0, len(revMap))
	for glyph, aliases := range revMap {
		if len(aliases) == 0 {
			continue
		}
		pairs = append(pairs, pair{unicode: glyph, shortcode: aliases[0]})
	}

	sort.Slice(pairs, func(i, j int) bool {
		if len(pairs[i].unicode) != len(pairs[j].unicode) {
			return len(pairs[i].unicode) > len(pairs[j].unicode)
		}
		return pairs[i].unicode < pairs[j].unicode
	})

	flat := make([]string, 0, len(pairs)*2)
	for _, p := range pairs {
		flat = append(flat, p.unicode, p.shortcode)
	}
	return strings.NewReplacer(flat...)
}

func (d *EmojiDecoder) DemojizeContent(content string) string {
	return d.replacer.Replace(content)
}
