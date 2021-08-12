package patterns

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPatterns_AddPattern(t *testing.T) {
	var p = NewMatcher()

	assert.Equal(t, false, p.MatchString("anything"))
}
func TestPatterns_TestingEmptyStringShouldReturnFalse(t *testing.T) {
	var p = NewMatcher()

	assert.Equal(t, false, p.MatchString(""))
}

func TestPatterns_ItShouldMatchPrefixes(t *testing.T) {
	var p = NewMatcher()
	p.AddStringPattern("name:*")

	assert.Equal(t, false, p.MatchString("name"))
	assert.Equal(t, true, p.MatchString("name:en"))
	assert.Equal(t, false, p.MatchString("anything"))
}

func TestPatterns_ItShouldMatchExactPatterns(t *testing.T) {
	var p = NewMatcher()
	p.AddStringPattern("name")

	assert.Equal(t, true, p.MatchString("name"))
	assert.Equal(t, false, p.MatchString("name:en"))
	assert.Equal(t, false, p.MatchString("anything"))
}

func TestPatterns_ItShouldMatchSuffixPatterns(t *testing.T) {
	var p = NewMatcher()
	p.AddStringPattern("*source")

	assert.Equal(t, true, p.MatchString("wikipedia:source"))
	assert.Equal(t, true, p.MatchString("ref:source"))
	assert.Equal(t, false, p.MatchString("anything"))
}

func setup(p *Patterns) {
	p.AddStringPattern("name")
	p.AddStringPattern("name:*")
	p.AddStringPattern("int_name")
	p.AddStringPattern("int_name:*")
	p.AddStringPattern("nat_name")
	p.AddStringPattern("nat_name:*")
	p.AddStringPattern("reg_name")
	p.AddStringPattern("reg_name:*")
	p.AddStringPattern("loc_name")
	p.AddStringPattern("loc_name:*")
	p.AddStringPattern("old_name")
	p.AddStringPattern("old_name:*")
	p.AddStringPattern("alt_name")
	p.AddStringPattern("alt_name:*")
	p.AddStringPattern("alt_name_*")
	p.AddStringPattern("official_name")
	p.AddStringPattern("official_name:*")
	p.AddStringPattern("place_name")
	p.AddStringPattern("place_name:*")
	p.AddStringPattern("short_name")
	p.AddStringPattern("short_name:*")
	p.AddStringPattern("brand")
	p.AddStringPattern("*wikipedia")
}

var _ bool

func BenchmarkExactMatch(b *testing.B) {
	b.ReportAllocs()
	var p = NewMatcher()
	setup(p)
	var text = []byte("name")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = p.Match(text)
	}
}

func BenchmarkPrefixMatch(b *testing.B) {
	b.ReportAllocs()
	var p = NewMatcher()
	setup(p)
	var text = []byte("name:en")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Match(text)
	}
}

func BenchmarkPrefixMatchUppercase(b *testing.B) {
	b.ReportAllocs()
	var p = NewMatcher()
	setup(p)
	var text = []byte("NAME:EN")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Match(text)
	}
}

func BenchmarkSuffixMatch(b *testing.B) {
	b.ReportAllocs()
	var p = NewMatcher()
	setup(p)
	var text = []byte("source:wikipedia")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Match(text)
	}
}

func BenchmarkNoMatch(b *testing.B) {
	b.ReportAllocs()
	var p = NewMatcher()
	setup(p)
	var text = []byte("qweoaihddnklc")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Match(text)
	}
}
