package regox

import (
	"strconv"
)

//Parse takes a string regex and parses it into a regex object
func Parse(regex string) Regex {
	return Regex{expression: regex, exprTree: cparse(regex)}
}

func cparse(regex string) consumer {
	return tparse(tokenize(regex))
}

func tparse(regex []string) consumer {
	return splitConcatenation(regex)
}

//SplitConcatenation takes a regex and separates it then sorts it into a concatenation
func splitConcatenation(regex []string) consumer {
	if len(regex) == 0 {
		return atom("")
	}
	str := regex
	var cons consumer
	parts := make([]consumer, 0)
	for len(str) > 0 {
		str, cons = splitRegex(str)
		parts = append([]consumer{cons}, parts...)
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return concat(parts...)
}

//SplitRegex splits a regex into a body and a tail, the tail being the trailing expression
func splitRegex(regex []string) ([]string, consumer) {
	lastToken := regex[len(regex)-1]
	if len(regex) == 1 {
		return make([]string, 0), splitSingular(regex[0])
	}
	if lastToken == ")" {
		body, tail := separens(regex, "(", ")")
		if parencontains(tail, "|") {
			return body, union(splitUnion(tail)...)
		}
		return body, capture(splitConcatenation(tail[1 : len(tail)-1]))
	}
	if lastToken == "}" {
		body, tail := separens(regex, "{", "}")
		nbody, repeater := splitRegex(body)
		if strcontains(tail[1], ',') {
			lower, upper := strSplit(tail[1], ',')
			l, _ := strconv.Atoi(lower)
			var u int
			if upper == "" {
				u = -1
			} else {
				u, _ = strconv.Atoi(upper)
			}
			return nbody, rangeRepeat(repeater, l, u)
		}
		repetitions, _ := strconv.Atoi(tail[1])
		return nbody, repeat(repeater, repetitions)
	}
	if lastToken == "]" {
		body, tail := separens(regex, "[", "]")
		setContents := abbreviate(tail)
		setTokens := setTokenize(setContents[1 : len(setContents)-1])
		if setTokens[0] == "^" {
			return body, negate(set(splitSet(setTokens)))
		}
		return body, set(splitSet(setTokens))
	}
	if lastToken == "*" {
		body, tail := splitRegex(regex[0 : len(regex)-1])
		return body, star(tail)
	}
	if lastToken == "+" {
		body, tail := splitRegex(regex[0 : len(regex)-1])
		return body, plus(tail)
	}
	if lastToken == "?" {
		body, tail := splitRegex(regex[0 : len(regex)-1])
		return body, option(tail)
	}
	return regex[0 : len(regex)-1], splitSingular(regex[len(regex)-1])
}

//SplitSingular takes an atomic regular expression and parses it
func splitSingular(regex string) consumer {
	if regex == "." {
		return any()
	}

	if regex[0] == '\\' {
		escChar := regex[1]
		if escChar == 'd' {
			return digit()
		}
		if escChar == '\\' {
			return backslash()
		}
		if escChar == 's' {
			return space()
		}
		if escChar == 't' {
			return tab()
		}
		if escChar == 'D' {
			return negate(digit())
		}
		if escChar == 'T' {
			return negate(tab())
		}
		if escChar == 'S' {
			return negate(space())
		}
		if escChar == 'w' {
			return word()
		}
		if escChar == 'W' {
			return negate(word())
		}
		return atom(string(escChar))
	}
	return atom(regex)
}

//SplitSet takes tokens (from a set tokenization) within a set and builds a consumer slice to generate a set Atom
func splitSet(regex []string) []consumer {
	cons := make([]consumer, 0)
	for _, token := range regex {
		var con consumer
		if token == "." {
			con = atom(".")
		} else if len(token) < 3 {
			con = splitSingular(token)
		} else {
			con = inRange(token[0], token[2])
		}
		cons = append(cons, con)
	}
	return cons
}

//SplitUnion takes a token section and splits it into an array of subexpressions, split by the pipe character
func splitUnion(regex []string) []consumer {
	tokBuffer := make([]string, 0)
	consumers := make([]consumer, 0)
	level := 0
	for _, token := range regex[1 : len(regex)-1] {
		if token == "(" {
			level++
		} else if token == ")" {
			level--
		}
		if level == 0 && token == "|" {
			consumers = append(consumers, splitConcatenation(tokBuffer))
			tokBuffer = make([]string, 0)
		} else {
			tokBuffer = append(tokBuffer, token)
		}
	}
	consumers = append(consumers, splitConcatenation(tokBuffer))
	return consumers
}

//	a-z-A-Z\\\\asA-zdf\\d.\\.-
//	a-z - A-Z \\\\ a s A-z d f \\d . \\. -

//setTokenize takes the contents of a set and tokenize it into elements
func setTokenize(s string) []string {
	buffer := ""
	tokens := make([]string, 0)
	if s[0] == '^' {
		tokens = append(tokens, "^")
		s = s[1:len(s)]
	}
	for _, char := range s {
		if len(buffer) == 0 {
			buffer += string(char)
		} else if len(buffer) == 1 {
			first := buffer[0]
			if first == '\\' {
				tokens = append(tokens, buffer+string(char))
				buffer = ""
			} else if char != '-' {
				tokens = append(tokens, buffer)
				buffer = string(char)
			} else {
				buffer += string(char)
			}
		} else if len(buffer) == 2 {
			tokens = append(tokens, buffer+string(char))
			buffer = ""
		}
	}
	if len(buffer) > 0 {
		tokens = append(tokens, buffer)
	}
	return tokens
}

//	aa\\\\bcd\\dasf(abc){2}de
//	aa \\ \\ bcd \\d asdf ( abc ) { 2 } de

//tokenize takes a regex and splits it into tokens
func tokenize(regex string) []string {
	buffer := ""
	isEscaping := false
	tokens := make([]string, 0)
	for _, c := range regex {
		if isEscaping {
			tokens = append(tokens, "\\"+string(c))
			isEscaping = false
			buffer = ""
		} else {
			if c == '\\' {
				if len(buffer) > 0 {
					tokens = append(tokens, buffer)
					buffer = ""
				}
				isEscaping = true
			} else if strcontains("()[]{}|.+*?", c) {
				if len(buffer) > 0 {
					tokens = append(tokens, buffer)
					buffer = ""
				}
				tokens = append(tokens, string(c))
			} else {
				buffer += string(c)
			}
		}
	}
	if len(buffer) > 0 {
		tokens = append(tokens, buffer)
		buffer = ""
	}
	return tokens
}

func deparens(tokens []string, opener, closer string) []string {
	level := 0
	for i := len(tokens) - 1; i >= 0; i-- {
		if tokens[i] == closer {
			level++
		} else if tokens[i] == opener {
			level--
		}
		if level == 0 {
			return tokens[i:len(tokens)]
		}
	}
	return tokens
}

//like deparens but returns the tokens split in two
func separens(tokens []string, opener, closer string) ([]string, []string) {
	parens := deparens(tokens, opener, closer)
	return tokens[0 : len(tokens)-len(parens)], parens
}

func strSplit(s string, splitter rune) (string, string) {
	buffer := ""
	for i, r := range s {
		if r == splitter {
			return s[0:i], s[i+1 : len(s)]
		}
		buffer += string(r)
	}
	return buffer, ""
}

func strcontains(s string, e rune) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func parencontains(tokens []string, elem string) bool {
	level := 0
	for _, token := range tokens {
		if token == "(" {
			level++
		} else if token == ")" {
			level--
		} else if level == 1 && token == elem {
			return true
		}
		if level == 0 {
			return false
		}
	}
	return false
}

func abbreviate(tokens []string) string {
	out := ""
	for _, token := range tokens {
		out += token
	}
	return out
}
