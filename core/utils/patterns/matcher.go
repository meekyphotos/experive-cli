package patterns

type patternNode struct {
	kids  map[byte]*patternNode
	final bool
	exact bool
}

type Patterns struct {
	normal  *patternNode
	reverse *patternNode
}

func NewMatcher() *Patterns {
	patterns := Patterns{}
	patterns.Init()
	return &patterns
}

func (p *Patterns) Init() {
	p.normal = &patternNode{kids: make(map[byte]*patternNode)}
	p.reverse = &patternNode{kids: make(map[byte]*patternNode)}
}

func (p *Patterns) AddStringPattern(pattern string) {
	p.AddPattern([]byte(pattern))
}

func (p *Patterns) AddPattern(pattern []byte) {
	if pattern[0] == '*' {
		p.storeReverse(pattern)
	} else {
		p.store(pattern)
	}
}

func (p *Patterns) MatchString(text string) bool {
	return p.Match([]byte(text))
}

func (p *Patterns) Match(text []byte) bool {
	if len(text) == 0 {
		return false
	}
	currFwKid := p.normal
	currBwKid := p.reverse
	var exist bool
	// check for prefixes
	var fwIdx = 0
	lenText := len(text)
	var bwIdx = lenText - 1
	for {
		if fwIdx != -1 && fwIdx < lenText {
			currFwKid, exist = currFwKid.kids[text[fwIdx]]
			if !exist {
				fwIdx = -1
			} else if currFwKid.final {
				return true
			} else if fwIdx == lenText-1 {
				return currFwKid.exact
			} else {
				fwIdx++
			}
		}

		if bwIdx >= 0 {
			currBwKid, exist = currBwKid.kids[text[bwIdx]]
			if !exist {
				bwIdx = -1
			} else if currBwKid.final {
				return true
			} else {
				bwIdx--
			}
		}

		if fwIdx == -1 && bwIdx == -1 {
			return false
		}
	}

}

type currToken func(pattern []byte) byte
type slicePattern func(pattern []byte) []byte

func normalGetter(pattern []byte) byte {
	return pattern[0]
}

func normalSlicer(pattern []byte) []byte {
	return pattern[1:]
}

func reverseGetter(pattern []byte) byte {
	return pattern[len(pattern)-1]
}

func reverseSlicer(pattern []byte) []byte {
	return pattern[:len(pattern)-1]
}

func storePattern(node *patternNode, pattern []byte, getter currToken, slicer slicePattern) {
	el := getter(pattern)
	if kid, exists := node.kids[el]; exists {
		if len(pattern) > 1 {
			storePattern(kid, slicer(pattern), getter, slicer)
		} else {
			kid.exact = true
		}
	} else {
		if el == '*' {
			node.final = true
		} else {
			newKid := &patternNode{kids: make(map[byte]*patternNode)}
			node.kids[el] = newKid
			if len(pattern) > 1 {
				storePattern(newKid, slicer(pattern), getter, slicer)
			} else {
				newKid.exact = true
			}
		}
	}
}

func (p *Patterns) store(pattern []byte) {
	storePattern(p.normal, pattern, normalGetter, normalSlicer)
}

func (p *Patterns) storeReverse(pattern []byte) {
	storePattern(p.reverse, pattern, reverseGetter, reverseSlicer)

}
