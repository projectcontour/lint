// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package messagefmt

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer for message formatting rules
var Analyzer = &analysis.Analyzer{
	Name:     "messagefmt",
	Doc:      "Check message formatting rules.",
	Run:      runMessageFmt,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func isException(word string) bool {
	var exceptions = map[string]struct{}{
		"xDS":  struct{}{},
		"gRPC": struct{}{},
	}

	_, ok := exceptions[word]
	return ok
}

func funcForCallExpr(pass *analysis.Pass, call *ast.CallExpr) (*types.Func, bool) {
	// Get the selector.
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}

	// Get the object this is calling.
	obj, ok := pass.TypesInfo.Uses[sel.Sel]
	if !ok {
		return nil, false
	}

	// The object should be a function, but check anyway.
	f, ok := obj.(*types.Func)
	return f, ok
}

// isFromPkg returns true if fun is in the package pkg.
func isFromPkg(fun *types.Func, pkg string) bool {
	// Calls to builtin types might not have a package.
	if fun.Pkg() != nil {
		return fun.Pkg().Path() == pkg
	}

	return false
}

func isFlagFnCall(fun *types.Func) bool {
	switch fun.Name() {
	case "Flag", "Command":
		return true
	default:
		return false
	}

}

func isLogFnCall(fun *types.Func) bool {
	names := []string{
		"Debug",
		"Error",
		"Fatal",
		"Panic",
		"Print",
		"Info",
		"Trace",
		"Warn",
		"Warning",
	}

	for _, n := range names {
		switch fun.Name() {
		case n, n + "f", n + "ln":
			return true
		}
	}

	return false
}

func getStringLiteralArgN(call *ast.CallExpr, argN int) *ast.BasicLit {
	if len(call.Args) <= argN {
		return nil
	}

	lit, ok := call.Args[argN].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return nil
	}

	return lit
}

func checkInitialLower(pass *analysis.Pass, lit *ast.BasicLit) {
	if lit == nil {
		return
	}

	words := strings.Fields(strings.Trim(lit.Value, "\"`"))
	first := words[0]

	// If the first word is all uppercase, it's an
	// initialism, so don't flag it.
	if first == strings.ToUpper(first) {
		return
	}

	// Decode the first UTF-8 character. Remember that
	// this is a string literal, so the first byte in
	// `lit.Value` will be the string literal quote.
	firstRune, _ := utf8.DecodeRuneInString(first)
	if unicode.IsUpper(firstRune) && !isException(first) {
		pass.Reportf(lit.Pos(), "message starts with uppercase: %s", lit.Value)
	}
}

func checkInitialUpper(pass *analysis.Pass, lit *ast.BasicLit) {
	if lit == nil {
		return
	}

	words := strings.Fields(strings.Trim(lit.Value, "\"`"))
	first := words[0]

	// If the first word is all uppercase, it's an
	// initialism, so don't flag it.
	if first == strings.ToUpper(first) {
		return
	}

	// Decode the first UTF-8 character. Remember that
	// this is a string literal, so the first byte in
	// `lit.Value` will be the string literal quote.
	firstRune, _ := utf8.DecodeRuneInString(first)
	if unicode.IsLower(firstRune) && !isException(first) {
		pass.Reportf(lit.Pos(), "message starts with lowercase: %s", lit.Value)
	}
}

func checkEndsWithoutPeriod(pass *analysis.Pass, lit *ast.BasicLit) {
	if lit == nil {
		return
	}

	value := strings.Trim(lit.Value, "\"`")

	if len(value) > 0 && value[len(value)-1] == '.' {
		pass.Reportf(lit.Pos(), "message must not end with a period: %s", lit.Value)
	}
}

func checkEndsWithPeriod(pass *analysis.Pass, lit *ast.BasicLit) {
	if lit == nil {
		return
	}

	value := strings.Trim(lit.Value, "\"`")

	if len(value) == 0 || value[len(value)-1] != '.' {
		pass.Reportf(lit.Pos(), "message must end with a period: %s", lit.Value)
	}
}

func runMessageFmt(pass *analysis.Pass) (interface{}, error) {
	i := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// We filter only function calls.
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	i.Preorder(nodeFilter, func(node ast.Node) {
		call := node.(*ast.CallExpr)

		// The object should be a function, but check anyway.
		fun, ok := funcForCallExpr(pass, call)
		if !ok {
			return
		}

		if isFromPkg(fun, "github.com/sirupsen/logrus") && isLogFnCall(fun) {
			checkInitialLower(pass, getStringLiteralArgN(call, 0))
			checkEndsWithoutPeriod(pass, getStringLiteralArgN(call, 0))
		}

		if isFromPkg(fun, "gopkg.in/alecthomas/kingpin.v2") && isFlagFnCall(fun) {
			checkInitialUpper(pass, getStringLiteralArgN(call, 1))
			checkEndsWithPeriod(pass, getStringLiteralArgN(call, 1))
		}

	})

	return nil, nil
}
