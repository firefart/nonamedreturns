package main

import "errors"

// allowed: pure documentation, explicit return, no reference
func explicitReturn(a, b int) (sum int) {
	return a + b
}

// reported: "sum" is assigned in the body
func assigned(a, b int) (sum int) { // want `named return "sum" with type "int" must not be referenced or used by a naked return`
	sum = a + b
	return sum
}

// reported: "sum" is read in the body
func read() (sum int) { // want `named return "sum" with type "int" must not be referenced or used by a naked return`
	if sum > 0 {
		return 1
	}
	return 0
}

// reported: address of "sum" is taken
func addressOf() (sum int) { // want `named return "sum" with type "int" must not be referenced or used by a naked return`
	p := &sum
	_ = p
	return 0
}

// reported: naked return implicitly uses "sum"
func nakedReturn() (sum int) { // want `named return "sum" with type "int" must not be referenced or used by a naked return`
	return
}

// reported: naked return implicitly uses every named result
func nakedReturnMulti() (a int, b string) { // want `named return "a" with type "int" must not be referenced or used by a naked return` `named return "b" with type "string" must not be referenced or used by a naked return`
	return
}

// reported: referenced inside a defer (no exemption in this mode)
func referencedInDefer() (err error) { // want `named return "err" with type "error" must not be referenced or used by a naked return`
	defer func() {
		err = nil
	}()
	return errors.New("x")
}

// allowed: "_" result name is skipped, explicit return
func underscoreResult() (_ int) {
	return 1
}

// allowed: a naked return only populates underscore results, which are never
// reported
func nakedReturnUnderscore() (_ int) {
	return
}

// reported: "err" is assigned by a range statement in the body
func rangeAssigned() (err error) { // want `named return "err" with type "error" must not be referenced or used by a naked return`
	for _, err = range []error{nil} {
		_ = err
	}
	return errors.New("x")
}

// allowed: func literal, explicit return, no reference
var goodLiteral = func() (sum int) {
	return 1
}

// reported: func literal with naked return
var badLiteral = func() (sum int) { // want `named return "sum" with type "int" must not be referenced or used by a naked return`
	return
}

type t struct{}

// allowed: method, explicit return, no reference
func (t) goodMethod() (sum int) {
	return 1
}

// reported: method referencing its named result
func (t) badMethod() (sum int) { // want `named return "sum" with type "int" must not be referenced or used by a naked return`
	sum = 1
	return sum
}

// mixed: only the referenced result is reported; the other uses explicit return
func mixed() (a int, b int) { // want `named return "a" with type "int" must not be referenced or used by a naked return`
	a = 1
	return a, 2
}

// reported: shared-type result group (single field, two names); only the
// referenced name is reported, both share the printed type.
func sharedGroup() (a, b int) { // want `named return "a" with type "int" must not be referenced or used by a naked return`
	a = 1
	return a, 2
}

// reported: "sum" is captured/referenced inside a non-defer closure
func referencedInClosure() (sum int) { // want `named return "sum" with type "int" must not be referenced or used by a naked return`
	g := func() int { return sum }
	return g()
}

// allowed: the naked return belongs to the nested closure, not to "sum".
// "sum" is never referenced and the outer function returns explicitly.
func nakedReturnInNestedClosure() (sum int) {
	f := func() { return }
	f()
	return 1
}
