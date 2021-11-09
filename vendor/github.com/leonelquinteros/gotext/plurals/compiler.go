/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

/*
 Package plurals is the pluralform compiler to get the correct translation id of the plural string
*/
package plurals

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type match struct {
	openPos  int
	closePos int
}

var pat = regexp.MustCompile(`(\?|:|\|\||&&|==|!=|>=|>|<=|<|%|\d+|n)`)

type testToken interface {
	compile(tokens []string) (test test, err error)
}

type cmpTestBuilder func(val uint32, flipped bool) test
type logicTestBuild func(left test, right test) test

var ternaryToken ternaryStruct

type ternaryStruct struct{}

func (ternaryStruct) compile(tokens []string) (expr Expression, err error) {
	main, err := splitTokens(tokens, "?")
	if err != nil {
		return expr, err
	}
	test, err := compileTest(strings.Join(main.Left, ""))
	if err != nil {
		return expr, err
	}
	actions, err := splitTokens(main.Right, ":")
	if err != nil {
		return expr, err
	}
	trueAction, err := compileExpression(strings.Join(actions.Left, ""))
	if err != nil {
		return expr, err
	}
	falseAction, err := compileExpression(strings.Join(actions.Right, ""))
	if err != nil {
		return expr, nil
	}
	return ternary{
		test:      test,
		trueExpr:  trueAction,
		falseExpr: falseAction,
	}, nil
}

var constToken constValStruct

type constValStruct struct{}

func (constValStruct) compile(tokens []string) (expr Expression, err error) {
	if len(tokens) == 0 {
		return expr, errors.New("got nothing instead of constant")
	}
	if len(tokens) != 1 {
		return expr, fmt.Errorf("invalid constant: %s", strings.Join(tokens, ""))
	}
	i, err := strconv.Atoi(tokens[0])
	if err != nil {
		return expr, err
	}
	return constValue{value: i}, nil
}

func compileLogicTest(tokens []string, sep string, builder logicTestBuild) (test test, err error) {
	split, err := splitTokens(tokens, sep)
	if err != nil {
		return test, err
	}
	left, err := compileTest(strings.Join(split.Left, ""))
	if err != nil {
		return test, err
	}
	right, err := compileTest(strings.Join(split.Right, ""))
	if err != nil {
		return test, err
	}
	return builder(left, right), nil
}

var orToken orStruct

type orStruct struct{}

func (orStruct) compile(tokens []string) (test test, err error) {
	return compileLogicTest(tokens, "||", buildOr)
}
func buildOr(left test, right test) test {
	return or{left: left, right: right}
}

var andToken andStruct

type andStruct struct{}

func (andStruct) compile(tokens []string) (test test, err error) {
	return compileLogicTest(tokens, "&&", buildAnd)
}
func buildAnd(left test, right test) test {
	return and{left: left, right: right}
}

func compileMod(tokens []string) (math math, err error) {
	split, err := splitTokens(tokens, "%")
	if err != nil {
		return math, err
	}
	if len(split.Left) != 1 || split.Left[0] != "n" {
		return math, errors.New("Modulus operation requires 'n' as left operand")
	}
	if len(split.Right) != 1 {
		return math, errors.New("Modulus operation requires simple integer as right operand")
	}
	i, err := parseUint32(split.Right[0])
	if err != nil {
		return math, err
	}
	return mod{value: uint32(i)}, nil
}

func subPipe(modTokens []string, actionTokens []string, builder cmpTestBuilder, flipped bool) (test test, err error) {
	modifier, err := compileMod(modTokens)
	if err != nil {
		return test, err
	}
	if len(actionTokens) != 1 {
		return test, errors.New("can only get modulus of integer")
	}
	i, err := parseUint32(actionTokens[0])
	if err != nil {
		return test, err
	}
	action := builder(uint32(i), flipped)
	return pipe{
		modifier: modifier,
		action:   action,
	}, nil
}

func compileEquality(tokens []string, sep string, builder cmpTestBuilder) (test test, err error) {
	split, err := splitTokens(tokens, sep)
	if err != nil {
		return test, err
	}
	if len(split.Left) == 1 && split.Left[0] == "n" {
		if len(split.Right) != 1 {
			return test, errors.New("test can only compare n to integers")
		}
		i, err := parseUint32(split.Right[0])
		if err != nil {
			return test, err
		}
		return builder(i, false), nil
	} else if len(split.Right) == 1 && split.Right[0] == "n" {
		if len(split.Left) != 1 {
			return test, errors.New("test can only compare n to integers")
		}
		i, err := parseUint32(split.Left[0])
		if err != nil {
			return test, err
		}
		return builder(i, true), nil
	} else if contains(split.Left, "n") && contains(split.Left, "%") {
		return subPipe(split.Left, split.Right, builder, false)
	}
	return test, errors.New("equality test must have 'n' as one of the two tests")

}

var eqToken eqStruct

type eqStruct struct{}

func (eqStruct) compile(tokens []string) (test test, err error) {
	return compileEquality(tokens, "==", buildEq)
}
func buildEq(val uint32, flipped bool) test {
	return equal{value: val}
}

var neqToken neqStruct

type neqStruct struct{}

func (neqStruct) compile(tokens []string) (test test, err error) {
	return compileEquality(tokens, "!=", buildNeq)
}
func buildNeq(val uint32, flipped bool) test {
	return notequal{value: val}
}

var gtToken gtStruct

type gtStruct struct{}

func (gtStruct) compile(tokens []string) (test test, err error) {
	return compileEquality(tokens, ">", buildGt)
}
func buildGt(val uint32, flipped bool) test {
	return gt{value: val, flipped: flipped}
}

var gteToken gteStruct

type gteStruct struct{}

func (gteStruct) compile(tokens []string) (test test, err error) {
	return compileEquality(tokens, ">=", buildGte)
}
func buildGte(val uint32, flipped bool) test {
	return gte{value: val, flipped: flipped}
}

var ltToken ltStruct

type ltStruct struct{}

func (ltStruct) compile(tokens []string) (test test, err error) {
	return compileEquality(tokens, "<", buildLt)
}
func buildLt(val uint32, flipped bool) test {
	return lt{value: val, flipped: flipped}
}

var lteToken lteStruct

type lteStruct struct{}

func (lteStruct) compile(tokens []string) (test test, err error) {
	return compileEquality(tokens, "<=", buildLte)
}
func buildLte(val uint32, flipped bool) test {
	return lte{value: val, flipped: flipped}
}

type testTokenDef struct {
	op    string
	token testToken
}

var precedence = []testTokenDef{
	{op: "||", token: orToken},
	{op: "&&", token: andToken},
	{op: "==", token: eqToken},
	{op: "!=", token: neqToken},
	{op: ">=", token: gteToken},
	{op: ">", token: gtToken},
	{op: "<=", token: lteToken},
	{op: "<", token: ltToken},
}

type splitted struct {
	Left  []string
	Right []string
}

// Find index of token in list of tokens
func index(tokens []string, sep string) int {
	for index, token := range tokens {
		if token == sep {
			return index
		}
	}
	return -1
}

// Split a list of tokens by a token into a splitted struct holding the tokens
// before and after the token to be split by.
func splitTokens(tokens []string, sep string) (s splitted, err error) {
	index := index(tokens, sep)
	if index == -1 {
		return s, fmt.Errorf("'%s' not found in ['%s']", sep, strings.Join(tokens, "','"))
	}
	return splitted{
		Left:  tokens[:index],
		Right: tokens[index+1:],
	}, nil
}

// Scan a string for parenthesis
func scan(s string) <-chan match {
	ch := make(chan match)
	go func() {
		depth := 0
		opener := 0
		for index, char := range s {
			switch char {
			case '(':
				if depth == 0 {
					opener = index
				}
				depth++
			case ')':
				depth--
				if depth == 0 {
					ch <- match{
						openPos:  opener,
						closePos: index + 1,
					}
				}
			}

		}
		close(ch)
	}()
	return ch
}

// Split the string into tokens
func split(s string) <-chan string {
	ch := make(chan string)
	go func() {
		s = strings.Replace(s, " ", "", -1)
		if !strings.Contains(s, "(") {
			ch <- s
		} else {
			last := 0
			end := len(s)
			for info := range scan(s) {
				if last != info.openPos {
					ch <- s[last:info.openPos]
				}
				ch <- s[info.openPos:info.closePos]
				last = info.closePos
			}
			if last != end {
				ch <- s[last:]
			}
		}
		close(ch)
	}()
	return ch
}

// Tokenizes a string into a list of strings, tokens grouped by parenthesis are
// not split! If the string starts with ( and ends in ), those are stripped.
func tokenize(s string) []string {
	/*
		TODO: Properly detect if the string starts with a ( and ends with a )
		and that those two form a matching pair.

		Eg: (foo) -> true; (foo)(bar) -> false;
	*/
	if len(s) == 0 {
		return []string{}
	}
	if s[0] == '(' && s[len(s)-1] == ')' {
		s = s[1 : len(s)-1]
	}
	ret := []string{}
	for chunk := range split(s) {
		if len(chunk) != 0 {
			if chunk[0] == '(' && chunk[len(chunk)-1] == ')' {
				ret = append(ret, chunk)
			} else {
				for _, token := range pat.FindAllStringSubmatch(chunk, -1) {
					ret = append(ret, token[0])
				}
			}
		} else {
			fmt.Printf("Empty chunk in string '%s'\n", s)
		}
	}
	return ret
}

// Compile a string containing a plural form expression to a Expression object.
func Compile(s string) (expr Expression, err error) {
	if s == "0" {
		return constValue{value: 0}, nil
	}
	if !strings.Contains(s, "?") {
		s += "?1:0"
	}
	return compileExpression(s)
}

// Check if a token is in a slice of strings
func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// Compiles an expression (ternary or constant)
func compileExpression(s string) (expr Expression, err error) {
	tokens := tokenize(s)
	if contains(tokens, "?") {
		return ternaryToken.compile(tokens)
	}
	return constToken.compile(tokens)
}

// Compiles a test (comparison)
func compileTest(s string) (test test, err error) {
	tokens := tokenize(s)
	for _, tokenDef := range precedence {
		if contains(tokens, tokenDef.op) {
			return tokenDef.token.compile(tokens)
		}
	}
	return test, errors.New("cannot compile")
}

func parseUint32(s string) (ui uint32, err error) {
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return ui, err
	}
	return uint32(i), nil
}
