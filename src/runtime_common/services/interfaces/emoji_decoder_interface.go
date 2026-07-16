package interfaces

// EmojiDecoderInterface maps emoji glyphs to shortcodes so messages
// round-trip through Slack/broker.
type EmojiDecoderInterface interface {
	// DemojizeContent maps emoji glyphs to shortcodes (e.g. "🔄" → ":arrows_counterclockwise:").
	DemojizeContent(content string) string
}
