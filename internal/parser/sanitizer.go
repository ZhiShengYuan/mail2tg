package parser

import (
	"github.com/microcosm-cc/bluemonday"
)

type Sanitizer struct {
	policy *bluemonday.Policy
}

func NewSanitizer() *Sanitizer {
	// Create a UGC (User Generated Content) policy
	p := bluemonday.UGCPolicy()

	// Allow common HTML elements for email viewing
	p.AllowElements("p", "div", "span", "br", "hr",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"strong", "b", "em", "i", "u", "s", "strike",
		"ul", "ol", "li",
		"blockquote", "pre", "code",
		"table", "thead", "tbody", "tr", "th", "td",
		"img")

	// Allow specific attributes
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "alt", "width", "height").OnElements("img")
	p.AllowAttrs("class", "id").Globally()
	p.AllowAttrs("colspan", "rowspan").OnElements("td", "th")

	// Require URLs to be http, https, or mailto
	p.RequireParseableURLs(true)
	p.AllowURLSchemes("http", "https", "mailto")

	// Remove scripts, forms, iframes, and other dangerous elements
	// (bluemonday does this by default)

	return &Sanitizer{policy: p}
}

func (s *Sanitizer) Sanitize(html string) string {
	return s.policy.Sanitize(html)
}
