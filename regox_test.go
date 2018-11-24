package regox

import (
	"fmt"
	"strconv"
	"testing"
)

func TestAtomicMatches(t *testing.T) {
	atomMatcher := atom("asdf")
	digitMatcher := digit()
	anyMatcher := any()
	spaceMatcher := space()
	tabMatcher := tab()
	bsMatcher := backslash()
	if !atomMatcher("asdf").Success {
		t.Error("'asdf' should have matched but didn't.")
	}
	if atomMatcher("asd").Success {
		t.Error("'asd' matched 'asdf' even though it doesn't have complete coverage")
	}
	if !atomMatcher("asdfa").Success {
		t.Error("asdfa should have matched but didn't")
	}
	if atomMatcher("").Success {
		t.Error("the empty string should not match, but did")
	}
	if !atom("")("").Success {
		t.Error("the empty string should only match an empty atom")
	}
	if !digitMatcher("8").Success {
		t.Error("8 should match but didn't")
	}
	if !digitMatcher("89").Success {
		t.Error("89 did not match even though it begins with a digit")
	}
	if digitMatcher("asd").Success {
		t.Error("non-digits matched but shouldn't have.")
	}
	if digitMatcher("").Success {
		t.Error("the empty string should not match a digit")
	}
	if !anyMatcher("heyo").Success {
		t.Error("'heyo' should have matched any, but didn't")
	}
	if !anyMatcher("8560").Success {
		t.Error("'8560' should have matched any, but didn't")
	}
	if anyMatcher("").Success {
		t.Error("'the empty string should not have matched any, but didn't")
	}
	if !spaceMatcher(" ").Success {
		t.Error("' ' should have matched but didn't.")
	}
	if !spaceMatcher("\n").Success {
		t.Error("'\\n' should have matched but didn't.")
	}
	if !spaceMatcher("\r").Success {
		t.Error("'\\r' should have matched but didn't.")
	}
	if !spaceMatcher("	").Success {
		t.Error("'	' should have matched but didn't.")
	}
	if spaceMatcher("a").Success {
		t.Error("'a' shouldn't have matched but did.")
	}
	if spaceMatcher("").Success {
		t.Error("'' shouldn't have matched but did.")
	}
	if !tabMatcher("	").Success {
		t.Error("'	' should have matched but didn't.")
	}
	if tabMatcher(" ").Success {
		t.Error("' ' shouldn't have matched but did.")
	}
	if !bsMatcher("\\").Success {
		t.Error("'\\' should have matched but didn't.")
	}
	if bsMatcher("").Success {
		t.Error("'' shouldn't have matched but did.")
	}
	if bsMatcher("a").Success {
		t.Error("'a' shouldn't have matched but did.")
	}
	if inRange('a', 'z')("").Success {
		t.Error("lambda shouldn't have matched but did.")
	}
	if word()("").Success {
		t.Error("lambda shouldn't match a word character")
	}
}

func TestRepeat(t *testing.T) {
	atomic := atom("asdf")
	repeater := repeat(atomic, 3)
	if !repeater("asdfasdfasdfas").Success {
		t.Error("3 repetitions and then some did not succeed.")
	}
	if repeater("").Success {
		t.Error("empty string succeeded but didn't have any repetitions.")
	}
	if repeater("asdfasdfas").Success {
		t.Error("2 repetitions and then a part of the next repetition succeeded but shouldn't have")
	}
}

func TestRangeRepeat(t *testing.T) {
	repeater := rangeRepeat(atom("a"), 2, 4)
	Assert(t, repeater("aa").Success, true)
	Assert(t, repeater("a").Success, false)
	Assert(t, repeater("aaaa").Success, true)
	Assert(t, repeater("aaaaa").Success, true)
	Assert(t, repeater("aaaaa").Coverage, "aaaa")
	Assert(t, repeater("aba").Success, false)
}

func TestConcat(t *testing.T) {
	atomic1 := atom("asdf")
	atomic2 := atom("jkl")
	conc := concat(atomic1, atomic2)
	if conc("").Success {
		t.Error("empty string passed concatenation but shouldn't have")
	}
	if conc("asdf").Success {
		t.Error("first regex passed concatenation but shouldn't have")
	}
	if conc("jkl").Success {
		t.Error("second regex passed concatenation but shouldn't have")
	}
	if !conc("asdfjkl").Success {
		t.Error("full string didn't pass concatenation but should have")
	}
	if !conc("asdfjklasdf").Success {
		t.Error("full string plus some didn't pass concatenation but should have")
	}
	if concat()("").Success {
		t.Error("concat of nothing should not occur.  As a result a concat of nothing should always fail")
	}
}

func TestOption(t *testing.T) {
	atomic1 := atom("asdf")
	atomic2 := atom("jkl")
	//asdfjkl or jkl should pass
	conc := concat(option(atomic1), atomic2)
	//lambda or asdf should pass
	concat2 := option(atomic1)
	//asdf or asdfjkl should pass
	concat3 := concat(atomic1, option(atomic2))
	if conc("").Success {
		t.Error("empty string passed first concat but shouldn't have")
	}
	if conc("asdf").Success {
		t.Error("optional regex passed first concat but shouldn't have")
	}
	if !conc("asdfjkl").Success {
		t.Error("full string didn't pass first concat but should have")
	}
	if !conc("jkl").Success {
		t.Error("minimal string didn't pass first concat but should have")
	}
	if !concat2("").Success {
		t.Error("lambda didn't pass the second concat but should have")
	}
	if !concat2("asdf").Success {
		t.Error("asdf didn't pass the second concat but should have")
	}
	if concat3("").Success {
		t.Error("lambda passed the third concat but shouldn't have")
	}
	if !concat3("asdf").Success {
		t.Error("asdf didn't pass the third concat but should have")
	}
	if !concat3("asdfjkl").Success {
		t.Error("asdfjkl didn't pass the third concat but should have")
	}
}

func TestStar(t *testing.T) {
	atomic := atom("asdf")
	atomic2 := atom("jkl")
	//covers (asdf)*
	cstar := star(atomic)
	//covers asdf(jkl)*
	star2 := concat(atomic, star(atomic2))

	if !cstar("").Success {
		t.Error("lambda didn't pass the first star but should have")
	}
	res := cstar("asdf")
	if !res.Success {
		t.Error("asdf didn't pass the first star but should have.")
	}
	if res.Coverage != "asdf" { //star is greedy
		t.Error("asdf wasn't fully captured but should have been")
	}

	if star2("").Success {
		t.Error("lambda passed the second star but shouldn't have")
	}
	if !star2("asdf").Success {
		t.Error("asdf didn't pass the second star but should have")
	}
	if !star2("asdfjkl").Success {
		t.Error("asdfjkl didn't pass the second star but should have")
	}
	if !star2("asdfjkljkl").Success {
		t.Error("asdfjkljkl didn't pass the second star but should have")
	}
	if star2("asdfjkljkl").Coverage != "asdfjkljkl" {
		t.Error("asdfjkljkl wasn't fully covered, instead coverage: " + star2("asdfjkljkl").Coverage)
	}
	if star2("asdfjkljklyayaya").Coverage != "asdfjkljkl" {
		t.Error("asdfjkljklyayaya had incorrect coverage: " + star2("asdfjkljklyayaya").Coverage)
	}
}

func TestPlus(t *testing.T) {
	atomic := atom("asdf")
	atomic2 := atom("jkl")
	//covers (asdf)*
	mplus := plus(atomic)
	//covers asdf(jkl)*
	plus2 := concat(atomic, plus(atomic2))

	if mplus("").Success {
		t.Error("lambda passed the first plus but shouldn't have")
	}
	res := mplus("asdf")
	if !res.Success {
		t.Error("asdf didn't pass the first plus but should have.")
	}
	if res.Coverage != "asdf" { //plus is greedy
		t.Error("asdf wasn't fully captured but should have been")
	}

	if plus2("").Success {
		t.Error("lambda passed the second plus but shouldn't have")
	}
	if plus2("asdf").Success {
		t.Error("asdf passed the second plus but shouldn't have")
	}
	if !plus2("asdfjkl").Success {
		t.Error("asdfjkl didn't pass the second plus but should have")
	}
	if !plus2("asdfjkljkl").Success {
		t.Error("asdfjkljkl didn't pass the second plus but should have")
	}
	if plus2("asdfjkljkl").Coverage != "asdfjkljkl" {
		t.Error("asdfjkljkl wasn't fully covered, instead coverage: " + plus2("asdfjkljkl").Coverage)
	}
	if plus2("asdfjkljklyayaya").Coverage != "asdfjkljkl" {
		t.Error("asdfjkljklyayaya had incorrect coverage: " + plus2("asdfjkljklyayaya").Coverage)
	}
}

func TestUnion(t *testing.T) {
	atom1 := atom("gaben")
	atom2 := atom("heidi")
	munion := union(atom1, atom2)
	emptyResult := munion("")
	firstResult := munion("gaben")
	secondResult := munion("heidi")
	compoundResult := munion("gabenheidi")
	garbageResult := munion("heyo I'm a rockstar")
	if emptyResult.Success {
		t.Error("lambda passed the union but shouldn't have")
	}
	if !firstResult.Success {
		t.Error("gaben didn't pass but should have")
	}
	if !secondResult.Success {
		t.Error("heidi didn't pass but should have")
	}
	if !compoundResult.Success {
		t.Error("compound result should have passed (as it starts with gaben) but didn't")
	}
	if compoundResult.Coverage != "gaben" {
		t.Error("compound result did not cover gaben.  Coverage: " + compoundResult.Coverage)
	}
	if garbageResult.Success {
		t.Error("garbage passed but shouldn't have")
	}
}

func TestCapture(t *testing.T) {
	atom1 := atom("a")
	atom2 := atom("bc")
	//(a*)bc
	capt := capture(star(atom1))
	expr := concat(capt, atom2)
	if expr("").Success {
		t.Error("lambda passed but shouldn't have")
	}
	res1 := expr("")
	res2 := expr("bc")
	res3 := expr("abc")

	if res1.Success {
		t.Error("lambda passed but shouldn't have")
	}
	if !res2.Success {
		t.Error("empty capture failed but should have passed")
	}
	if len(res2.Captures) == 0 {
		t.Error("captures of empty capture does not contain empty string but should")
	}
	if res2.Captures[0] != "" {
		t.Error("captures of empty capture does not contain empty string, but rather " + res2.Captures[0])
	}
	if len(res3.Captures) != 1 {
		t.Error("captures of abc should just contain a, but instead has " + strconv.Itoa(len(res3.Captures)) + " elements")
	}

	compCapt := capture(concat(atom1, option(capture(atom2))))
	res1 = compCapt("")
	res2 = compCapt("a")
	res3 = compCapt("abc")

	if res1.Success {
		t.Error("lambda passed comp capt but shouldn't have")
	}
	if !res2.Success {
		t.Error("a should pass comp capt but doesn't")
	}
	if len(res2.Captures) != 1 {
		t.Error("a should have one capture group but instead has " + strconv.Itoa(len(res2.Captures)))
	}
	if res2.Captures[0] != "a" {
		t.Error("the first capture group of a should be a but is instead " + res2.Captures[0])
	}
	if len(res3.Captures) != 2 {
		t.Error("abc should have 2 capture groups but instead has " + strconv.Itoa(len(res3.Captures)))
	}
	if res3.Captures[0] != "abc" || res3.Captures[1] != "bc" {
		t.Error(fmt.Sprint("abc should have captures [abc, bc] but instead has ", res3.Captures))
	}

	crazyCapt := capture(repeat(capture(atom("abc")), 3))
	res := crazyCapt("abcabcabc")
	if res.Captures[0] != "abcabcabc" || res.Captures[1] != "abc" {
		t.Error(fmt.Sprint("abcabcabc should have captures [abcabcabc, abc] but instead has ", res.Captures))
	}
}

func TestNegate(t *testing.T) {
	negater := negate(space())
	Assert(t, negater("a").Success, true)
	Assert(t, negater(" ").Success, false)
	Assert(t, negater("").Success, false)
}

func TestParse(t *testing.T) {
	rgx := Parse("a+")
	ergx := plus(atom("a"))
	success2 := ergx("aa").Success
	success := rgx.Match("aa").Success
	if success != success2 {
		t.Error("lambda matched but shouldn't have")
	}
	phones := Parse("(\\(?\\d{3}\\)?)-?\\d{3}-?\\d{4}")
	res := phones.Match("(781)-729-5778")
	//manMatch := concat(capture(concat(option(atom("(")))))
	if !res.Success {
		t.Error("(781)-729-5778 should have matched but didn't")
	} else {
		Assert(t, res.Captures[1], "(781)")
	}
	rgx = Parse("")
	success = rgx.Match("").Success
	if !success {
		t.Error("empty string failed to parse")
	}
}

func TestSetTokenize(t *testing.T) {
	Assert(t, fmt.Sprint(setTokenize("a-z-A-Z\\\\asA-zdf\\d.\\.-")), fmt.Sprint("[a-z - A-Z \\\\ a s A-z d f \\d . \\. -]"))
}

func TestManyRegexes(t *testing.T) {
	r1 := Parse("c*")
	success := r1.Match("").Success
	success2 := r1.Match("c").Success
	success3 := r1.Match("ccc").Success
	if !(success || success2 || success3) {
		t.Error("one of c* matches failed")
	}

	r1 = Parse("a{2}")
	success = r1.Match("aa").Success
	Assert(t, success, true)

	r1 = Parse("\\\\{3}")
	success = r1.Match("\\\\\\").Success
	Assert(t, success, true)

	r1 = Parse("[Gg]ab(e|riel)")
	success = r1.Match("Gabe").Success
	Assert(t, success, true)
	res := r1.Match("Gabriel")
	Assert(t, res.Success, true)
	Assert(t, res.Captures[1], "riel")
	success = r1.Match("Gabnl").Success
	Assert(t, success, false)
	success = r1.Match("gabe").Success
	Assert(t, success, true)
	success = r1.Match("gabriel").Success
	Assert(t, success, true)
	success = r1.Match("sabe").Success
	Assert(t, success, false)

	r1 = Parse("[a-c]{3}")
	success = r1.Match("acb").Success
	Assert(t, success, true)
	success = r1.Match("adb").Success
	Assert(t, success, false)

	r1 = Parse("[a-c]")
	success = r1.Match("").Success
	Assert(t, success, false)

	r1 = Parse("[^a-c]")
	success = r1.Match("d").Success
	Assert(t, success, true)
	success = r1.Match("c").Success
	Assert(t, success, false)

	r1 = Parse(".[.]\\s\\t\\D\\T\\S")
	success = r1.Match("a. 	fgh").Success
	Assert(t, success, true)

	r1 = Parse("(asdf|h(i|j)k)\\w\\W")
	success = r1.Match("hjka9").Success
	Assert(t, success, true)

	r1 = Parse("a{2,5}")
	Assert(t, r1.Matches("aa"), true)
	Assert(t, r1.Matches("aaaaa"), true)
	Assert(t, r1.Matches("a"), false)
	Assert(t, r1.Matches("aaaaaa"), true)

	r1 = Parse("a{2,}")
	Assert(t, r1.Matches("aa"), true)
	Assert(t, r1.Matches("aaaaa"), true)
	Assert(t, r1.Matches("a"), false)
	Assert(t, r1.Matches("aaaaaa"), true)
}

func TestMatchAll(t *testing.T) {
	r := Parse("[Gg]ab(e|riel)")
	results, indices := r.MatchAll("Gabe gabriel Gabriel")
	Assert(t, len(results), 3)
	Assert(t, indices[0], 0)
	Assert(t, indices[1], 5)
	Assert(t, indices[2], 13)
}

func Assert(t *testing.T, value, expected interface{}) {
	if value != expected {
		t.Error(fmt.Sprint("expected ", expected, " but got ", value))
	}
}
