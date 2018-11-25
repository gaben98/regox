## Regox

Regox is a custom-built regex engine implemented in GOLANG.  Regox builds expression trees for evaluating regular expression matching, heavily levying GO's Closure support.  

## Getting started

Regox can be added to your project from command line with 
`go get github.com/gaben98/regox`
then can be imported in your code with
`import "github.com/gaben98/regox"`

## Use

A `Regex` object can be created using `regox.Parse(regex string)`

A `Regex` object can call `Match(s string)` to check if string s matches the regular expression.  This returns a `RegResult` object, which has three properties:
- `Success` is a `bool` whether or not the string matched the regular expression
- `Captures` is a `[]string` containing the capture groups from the match, and
- `Coverage` is a `string` containing the substring matched by the regular expression.

A `Regex` object can call `Match(s string)` which just returns a `bool` of whether the string matched the regular expression

A `Regex` object can call `MatchAll(s string)` which returns a `([]RegResult, []int)` that holds the `RegResult` and index of each substring match within `s`
