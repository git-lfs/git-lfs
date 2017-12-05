package githistory

import (
	"testing"
	"text/template"
)

func TestMessageRewriterRewriteRewritesMessagesWithoutExtraHeaders(t *testing.T) {
	r := NewMessageRewriter(template.Must(template.New("test", "rewrite")))
}
