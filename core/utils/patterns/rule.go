package patterns

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type StyleJson struct {
	Keys   []string          `json:"keys"`
	Values map[string]string `json:"values"`
}
type Rule struct {
	patterns             *Patterns
	defaultOutcome       int
	defaultOutcomeExists bool
	outcomeByValue       map[string]int
}

func NewRule(json StyleJson) *Rule {
	r := Rule{}
	r.patterns = NewMatcher()
	for _, v := range json.Keys {
		r.patterns.AddStringPattern(v)
	}
	r.outcomeByValue = make(map[string]int)
	for k, v := range json.Values {
		r.outcomeByValue[k] = parse(v)
	}
	r.defaultOutcome, r.defaultOutcomeExists = r.outcomeByValue[""]
	return &r
}

const (
	SfMain          = 1 << 0
	SfMainNamed     = 1 << 1
	SfMainNamedKey  = 1 << 2
	SfMainFallback  = 1 << 3
	SfMainOperator  = 1 << 4
	SfName          = 1 << 5
	SfRef           = 1 << 6
	SfAddress       = 1 << 7
	SfHouse         = 1 << 8
	SfPostcode      = 1 << 9
	SfCountry       = 1 << 10
	SfExtra         = 1 << 11
	SfInterpolation = 1 << 12
)

func parse(s string) int {
	var curr = s
	var out = 0
forLoop:
	for {
		next := strings.IndexRune(curr, ',')
		var token string
		if next == -1 {
			token = curr
		} else {
			token = curr[:next]
		}
		switch token {
		case "main":
			out |= SfMain
		case "with_name":
			out |= SfMainNamed
		case "with_name_key":
			out |= SfMainNamedKey
		case "fallback":
			out |= SfMainFallback
		case "operator":
			out |= SfMainOperator
		case "name":
			out |= SfName
		case "ref":
			out |= SfRef
		case "address":
			out |= SfAddress
		case "house":
			out |= SfHouse
		case "postcode":
			out |= SfPostcode
		case "country":
			out |= SfCountry
		case "extra":
			out |= SfExtra
		case "interpolation":
			out |= SfInterpolation
		case "skip":
			out = 0
			break forLoop
		}
		if next == -1 {
			break forLoop
		}
		curr = curr[next+1:]
	}
	return out
}

func (r *Rule) Matches(key []byte) bool {
	return r.patterns.Match(key)
}

func (r *Rule) GetFlag(value string) int {
	v, ok := r.outcomeByValue[value]
	if ok {
		return v
	}
	return r.defaultOutcome
}

func ParseRules(path string) ([]StyleJson, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data := make([]StyleJson, 0, 10)
	err = json.Unmarshal(file, &data)
	return data, err
}

func ParseStringRules(jsonData []byte) ([]StyleJson, error) {
	data := make([]StyleJson, 0, 10)
	err := json.Unmarshal(jsonData, &data)
	return data, err
}
