package wildcard

const (
	_normal     = iota
	_all        // *
	_any        // ?
	_setSymbol  // []
	_rangSymbol // [a-b]
	_negSymbol  // [^a]
)

type item struct {
	character byte
	set       map[byte]bool
	typeCode  int
}

func (i *item) contains(c byte) bool {
	if i.typeCode == _setSymbol {
		_, ok := i.set[c]
		return ok
	} else if i.typeCode == _rangSymbol {
		if _, ok := i.set[c]; ok {
			return true
		}
		var (
			_min uint8 = 255
			_max uint8 = 0
		)
		for k := range i.set {
			if _min > k {
				_min = k
			}
			if _max < k {
				_max = k
			}
		}
		return c >= _min && c <= _max
	} else {
		_, ok := i.set[c]
		return !ok
	}
}

// Pattern represents a wildcard pattern
type Pattern struct {
	items []*item
}

// CompilePattern convert wildcard string to Pattern
func CompilePattern(src string) *Pattern {
	items := make([]*item, 0)
	escape := false
	inSet := false
	var set map[byte]bool
	for _, v := range src {
		c := byte(v)
		if escape {
			items = append(items, &item{typeCode: _normal, character: c})
			escape = false
		} else if c == '*' {
			items = append(items, &item{typeCode: _all})
		} else if c == '?' {
			items = append(items, &item{typeCode: _any})
		} else if c == '\\' {
			escape = true
		} else if c == '[' {
			if !inSet {
				inSet = true
				set = make(map[byte]bool)
			} else {
				set[c] = true
			}
		} else if c == ']' {
			if inSet {
				inSet = false
				typeCode := _setSymbol
				if _, ok := set['-']; ok {
					typeCode = _rangSymbol
					delete(set, '-')
				}
				if _, ok := set['^']; ok {
					typeCode = _negSymbol
					delete(set, '^')
				}
				items = append(items, &item{typeCode: typeCode, set: set})
			} else {
				items = append(items, &item{typeCode: _normal, character: c})
			}
		} else {
			if inSet {
				set[c] = true
			} else {
				items = append(items, &item{typeCode: _normal, character: c})
			}
		}
	}
	return &Pattern{
		items: items,
	}
}

// IsMatch returns whether the given string matches pattern
func (p *Pattern) IsMatch(s string) bool {
	if len(p.items) == 0 {
		return len(s) == 0
	}
	m := len(s)
	n := len(p.items)
	table := make([][]bool, m+1)
	for i := 0; i < m+1; i++ {
		table[i] = make([]bool, n+1)
	}
	table[0][0] = true
	for j := 1; j < n+1; j++ {
		table[0][j] = table[0][j-1] && p.items[j-1].typeCode == _all
	}
	for i := 1; i < m+1; i++ {
		for j := 1; j < n+1; j++ {
			if p.items[j-1].typeCode == _all {
				table[i][j] = table[i-1][j] || table[i][j-1]
			} else {
				table[i][j] = table[i-1][j-1] &&
					(p.items[j-1].typeCode == _any ||
						(p.items[j-1].typeCode == _normal && uint8(s[i-1]) == p.items[j-1].character) ||
						(p.items[j-1].typeCode >= _setSymbol && p.items[j-1].contains(s[i-1])))
			}
		}
	}
	return table[m][n]
}
