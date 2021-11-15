/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package plurals

// Expression is a plurals expression. Eval evaluates the expression for
// a given n value. Use plurals.Compile to generate Expression instances.
type Expression interface {
	Eval(n uint32) int
}

type constValue struct {
	value int
}

func (c constValue) Eval(n uint32) int {
	return c.value
}

type test interface {
	test(n uint32) bool
}

type ternary struct {
	test      test
	trueExpr  Expression
	falseExpr Expression
}

func (t ternary) Eval(n uint32) int {
	if t.test.test(n) {
		if t.trueExpr == nil {
			return -1
		}
		return t.trueExpr.Eval(n)
	}
	if t.falseExpr == nil {
		return -1
	}
	return t.falseExpr.Eval(n)
}
