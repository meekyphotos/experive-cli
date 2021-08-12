package patterns

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSkipAllSources(t *testing.T) {
	var sj = StyleJson{
		Keys:   []string{"*source"},
		Values: map[string]string{"": "skip"},
	}
	var r = NewRule(sj)

	outcome, isValid := r.FindFlag([]byte("ref:source"), "wikipedia")
	assert.Equal(t, 0, outcome)
	assert.Equal(t, true, isValid)

	_, isValid = r.FindFlag([]byte("name"), "some name")
	assert.Equal(t, false, isValid)
}

func TestPutSomeValuesInExtra(t *testing.T) {
	var sj = StyleJson{
		Keys: []string{"name:prefix",
			"name:suffix",
			"name:prefix:*",
			"name:suffix:*",
			"name:etymology",
			"name:signed",
			"name:botanical",
			"wikidata",
			"*:wikidata"},
		Values: map[string]string{"": "extra"},
	}
	var r = NewRule(sj)

	outcome, isValid := r.FindFlag([]byte("name:suffix"), "Inc")
	assert.Equal(t, SfExtra, outcome)
	assert.Equal(t, true, isValid)
}

func TestMultiTypeMatching(t *testing.T) {
	var sj = StyleJson{
		Keys:   []string{"addr:housename"},
		Values: map[string]string{"": "name,house"},
	}
	var r = NewRule(sj)

	outcome, isValid := r.FindFlag([]byte("addr:housename"), "Villa")
	assert.True(t, isValid)
	assert.Equal(t, SfName, outcome&SfName)
	assert.Equal(t, SfHouse, outcome&SfHouse)
	assert.NotEqual(t, SfMain, outcome&SfMain)
}

func TestDifferentMatchesBasedOnValue(t *testing.T) {
	var sj = StyleJson{
		Keys:   []string{"emergency"},
		Values: map[string]string{"": "main", "yes": "skip", "no": "skip"},
	}
	var r = NewRule(sj)

	outcome, isValid := r.FindFlag([]byte("emergency"), "yes")
	assert.True(t, isValid)
	assert.Equal(t, 0, outcome)

	outcome, isValid = r.FindFlag([]byte("emergency"), "a value")
	assert.True(t, isValid)
	assert.Equal(t, SfMain, outcome&SfMain)

}

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func TestParseRules(t *testing.T) {
	r, err := ParseRules("import-full.style")
	assert.Nil(t, err)
	assert.Equal(t, 29, len(r))
}

func TestMultipleCommas(t *testing.T) {
	r := parse("main,with_name,fallback")
	assert.Equal(t, SfMain|SfMainNamed|SfMainFallback, r)
}
