package regox

//Regex holds the expression to be used in matching
type Regex struct {
	expression string
	exprTree   consumer
}

//RegResult holds the result of a regex match
type RegResult struct {
	Success  bool     //did this consumption succeed?
	Captures []string //what capture groups are there from this consumption?
	Coverage string   //how much of the input string does this consumption cover?
}

//consumer is an expression node in the regular expression tree used for evaluating matches.  It takes a string to match and returns a result
type consumer func(string) RegResult

//Match takes a string s and returns if it matches, as well as a slice of capture groups
func (regex *Regex) Match(s string) RegResult {
	captures := make([]string, 1)
	captures[0] = s
	res := regex.exprTree(s)
	res.Captures = append(captures, res.Captures...)
	return res
}

//Matches returns whether a given string s matches this regex
func (regex *Regex) Matches(s string) bool {
	return regex.exprTree(s).Success
}

//MatchAll returns all matches within a given string
func (regex *Regex) MatchAll(s string) []RegResult {
	matches := make([]RegResult, 0)
	i := 0
	for i < len(s) {
		res := regex.exprTree(s[i:len(s)])
		if res.Success {
			i += len(res.Coverage)
			matches = append(matches, res)
		} else {
			i++
		}
	}
	return matches
}

//I need to break up a regex into a composition of atomic regexes and operations
//(\(?\d{3}\)?)* becomes star(capture(concat(option(atom("(")), repeat(digit(), 3), option(atom(")")))))
//(asdf)? becomes option(capture(atom("asdf")))
//each atomic operation should return the number of characters consumed, -1 if match fails, and also a slice of thus captured groups to propagate upwards
//(54(63)
//star(capture(concat(option(atom("(")), repeat(digit(), 3), option(atom(")")))))
//star(capture(concat(option(success), repeat(yup, 3), option(failure))))
//star(capture(concat(success, failure, success)))
//star(capture(failure))
//star(failure)
//success

//atomics

//Atom matches a continuous sequence of explicit characters.
func atom(matcher string) consumer {
	return func(input string) RegResult {
		if len(input) < len(matcher) {
			return failure()
		}
		if input[0:len(matcher)] == matcher {
			return result(true, make([]string, 0), matcher)
		}
		return failure()
	}
}

//Word matches any alphabetical character
func word() consumer {
	return func(input string) RegResult {
		if len(input) == 0 {
			return failure()
		}
		if input[0] >= 'A' && input[0] <= 'z' {
			return result(true, make([]string, 0), input[0:1])
		}
		return failure()
	}
}

//Digit matches a singular digit of any value 0-9
func digit() consumer {
	return func(input string) RegResult {
		if len(input) == 0 {
			return failure()
		}
		if input[0] >= '0' && input[0] <= '9' {
			return result(true, make([]string, 0), input[0:1])
		}
		return failure()
	}
}

//Any matches a wild card
func any() consumer {
	return func(input string) RegResult {
		if len(input) > 0 {
			return result(true, make([]string, 0), input[0:1])
		}
		return failure()
	}
}

//Backslash matches a backslash literal
func backslash() consumer {
	return func(input string) RegResult {
		if input == "" {
			return failure()
		}
		if input[0] == '\\' {
			return result(true, make([]string, 0), input[0:1])
		}
		return failure()
	}
}

//Space matches a space, newline or tab character
func space() consumer {
	return func(input string) RegResult {
		if input == "" {
			return failure()
		}
		char := input[0]
		if strcontains("	\r \n", rune(char)) {
			return result(true, make([]string, 0), input[0:1])
		}
		return failure()
	}
}

//Tab matches just the tab character
func tab() consumer {
	return func(input string) RegResult {
		char := input[0]
		if char == '	' {
			return result(true, make([]string, 0), input[0:1])
		}
		return failure()
	}
}

//operations
//? character: string may or may not contain the contained regex.  If RegResult fails, then the string doesn't match the interior regex, and nothing should be consumed.  Otherwise RegResult should consume characters and propagate captures

//Negate matches a single character that doesn't match the contained expression
func negate(cons consumer) consumer {
	return func(input string) RegResult {
		if input == "" {
			return failure()
		}
		res := cons(input)
		if !res.Success {
			return result(true, make([]string, 0), input[0:1])
		}
		return failure()
	}
}

//Set matches any char within the string chars
func set(cons []consumer) consumer {
	return func(input string) RegResult {
		if input == "" {
			return failure()
		}
		for _, con := range cons {
			res := con(input)
			if res.Success {
				return result(true, make([]string, 0), input[0:1])
			}
		}
		return failure()
	}
}

//Range represents a character in between the lower and upper rune.  Only used in a set.
func inRange(lower, upper byte) consumer {
	return func(input string) RegResult {
		if input == "" {
			return failure()
		}
		if input[0] >= lower && input[0] <= upper {
			return result(true, make([]string, 0), input[0:1])
		}
		return failure()
	}
}

//Option matches ?, either 0 or one of the internal expression
func option(cons consumer) consumer {
	return func(input string) RegResult {
		res := cons(input)
		if !res.Success {
			return result(true, make([]string, 0), "")
		}
		return result(true, res.Captures, res.Coverage)
	}
}

//Repeat matches .{5}, repeats of the internal expression
func repeat(cons consumer, repetitions int) consumer {
	return func(input string) RegResult {
		coverage := ""
		captures := make([]string, 0)
		for reps := repetitions; reps > 0; reps-- {
			res := cons(input[len(coverage):len(input)])
			if !res.Success {
				return failure()
			}
			coverage += res.Coverage
			captures = append(captures, res.Captures...)
		}
		return result(true, captures, coverage)
	}
}

//RangeRepeat matches a subexpression repeated anywhere from minReps to maxReps times
func rangeRepeat(cons consumer, minReps, maxReps int) consumer {
	return func(input string) RegResult {
		coverage := ""
		captures := make([]string, 0)
		repsComplete := 0
		reps := maxReps
		if reps == -1 {
			reps = len(input)
		}
		for ; reps > 0; reps-- {
			res := cons(input[len(coverage):len(input)])
			if !res.Success {
				break
			}
			coverage += res.Coverage
			repsComplete++
			captures = append(captures, res.Captures...)
		}
		if repsComplete >= minReps && (maxReps == -1 || repsComplete <= maxReps) {
			return result(true, captures, coverage)
		}
		return failure()
	}
}

//Star matches 0 or more of the internal expression
func star(cons consumer) consumer {
	return func(input string) RegResult {
		instances := 0 //how many times was the consumer satisfied?
		index := 0     //the index to inspect the input from
		coverage := ""
		captures := make([]string, 0)
		for index < len(input) {
			res := cons(input[index:len(input)])
			if res.Success {
				instances++
				coverage += res.Coverage
				index += len(res.Coverage)
			} else {
				break
			}
		}
		return result(true, captures, coverage)
	}
}

//Plus matches 1 or more of the contained expression
func plus(cons consumer) consumer {
	return func(input string) RegResult {
		instances := 0 //how many times was the consumer satisfied?
		index := 0     //the index to inspect the input from
		coverage := ""
		captures := make([]string, 0)
		for index < len(input) {
			res := cons(input[index:len(input)])
			if res.Success {
				instances++
				coverage += res.Coverage
				captures = append(captures, res.Captures...)
				index += len(res.Coverage)
			} else {
				break
			}
		}
		if instances > 0 {
			return result(true, captures, coverage)
		}
		return failure()
	}
}

//Concat matches sequential regular expressions; ABC is concat(A,B,C)
func concat(consumers ...consumer) consumer {
	return func(input string) RegResult {
		coverage := ""
		captures := make([]string, 0)
		bigstr := input
		if len(consumers) == 0 {
			return failure()
		}
		for _, cons := range consumers {
			res := cons(bigstr)
			bigstr = bigstr[len(res.Coverage):len(bigstr)]
			if !res.Success {
				return failure()
			}
			coverage += res.Coverage
			captures = append(captures, res.Captures...)
		}
		return result(true, captures, coverage)
	}
}

//Union matches: (A|B|C) is union(A,B,C), result placed in capture group
func union(consumers ...consumer) consumer {
	return func(input string) RegResult {
		for _, cons := range consumers {
			res := cons(input)
			if res.Success {
				return result(true, append(res.Captures, res.Coverage), res.Coverage)
			}
		}
		return failure()
	}
}

//Capture captures information to be propagated upwards for analysis
func capture(cons consumer) consumer {
	return func(input string) RegResult {
		res := cons(input)
		if !res.Success {
			return failure()
		}
		return result(true, append([]string{res.Coverage}, res.Captures...), res.Coverage)
	}
}

//util
func result(success bool, captures []string, coverage string) RegResult {
	return RegResult{Success: success, Captures: captures, Coverage: coverage}
}

func failure() RegResult {
	return result(false, nil, "")
}
