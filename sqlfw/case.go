package sqlfw

import (
	"slices"
	"strings"
	"unicode"
)

func toSnake(s string) string {
	tzer := caseTokenizer{d: []byte(s), i: 0}
	b := strings.Builder{}
	first := true
	for i, j, ok := tzer.next(); ok; i, j, ok = tzer.next() {
		if !first {
			b.WriteByte('_')
		}
		b.WriteString(strings.ToLower(s[i:j]))
		first = false
	}
	return b.String()
}

type caseTokenizer struct {
	d []byte
	i int
}

func (t *caseTokenizer) next() (int, int, bool) {
	tokenBegin := optional[int]{}
	for ; t.i < len(t.d); t.i++ {
		if isStatementDelim(t.d[t.i]) {
			if tokenBegin.HasValue() {
				return tokenBegin.Value(), t.i, true
			}
			return 0, 0, false
		}
		if t.isTokenBegin(t.i-1, t.i) {
			if tokenBegin.HasValue() {
				return tokenBegin.Value(), t.i, true
			}
			tokenBegin.SetValue(t.i)
		} else if tokenBegin.HasValue() && t.isTokenEnd(t.i-1, t.i) {
			return tokenBegin.Value(), t.i, true
		}
	}
	if tokenBegin.HasValue() {
		return tokenBegin.Value(), len(t.d), true
	}
	return 0, 0, false
}

func (t *caseTokenizer) isTokenBegin(i, j int) bool {
	if isTokenDelim(t.d[j]) {
		return false
	}
	if i == -1 {
		return true
	}
	if isTokenDelim(t.d[i]) {
		return true
	}
	return isLower(t.d[i]) && isUpper(t.d[j])
}

func (t *caseTokenizer) isTokenEnd(i, j int) bool {
	if isTokenDelim(t.d[j]) {
		return true
	}
	return isLower(t.d[i]) && isUpper(t.d[j])
}

func isTokenDelim(b byte) bool {
	return slices.Contains(tokenDelimChars, b)
}

func isStatementDelim(b byte) bool {
	return !slices.Contains(tokenDelimChars, b) && unicode.IsSpace(rune(b))
}

func isUpper(b byte) bool {
	return 'A' <= b && b <= 'Z'
}

func isLower(b byte) bool {
	return 'a' <= b && b <= 'z'
}

var tokenDelimChars = []byte{' ', '_', '-', '.'}
