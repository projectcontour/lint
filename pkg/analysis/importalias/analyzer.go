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

package importalias

import (
	"fmt"
	"go/ast"
	"go/types"

	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer for import aliases
var Analyzer = &analysis.Analyzer{
	Name:     "importalias",
	Doc:      "Checks import aliases have consistent names",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{ // filter needed nodes: visit only them
		(*ast.ImportSpec)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		importStmt := node.(*ast.ImportSpec)

		if importStmt.Name == nil {
			return
		}

		alias := importStmt.Name.Name
		if alias == "" {
			return
		}

		if strings.HasPrefix(alias, "_") {
			return // Used by go test and for auto-includes, not a conflict.
		}

		aliasSlice := strings.Split(alias, "_")

		originalImportPath, _ := strconv.Unquote(importStmt.Path.Value)
		// replace all separators with `/` for normalization
		replacer := strings.NewReplacer("_", "/", ".", "/", "-", "")
		path := replacer.Replace(originalImportPath)
		// omit the domain name in path
		pathSlice := strings.Split(path, "/")[1:]

		if !checkVersion(aliasSlice[len(aliasSlice)-1], pathSlice) {
			applicableAlias := getAliasFix(pathSlice)
			_, versionIndex := packageVersion(pathSlice)
			pass.Report(analysis.Diagnostic{
				Pos: node.Pos(),
				Message: fmt.Sprintf("version %q not specified in alias %q for import path %q may replace %q with %q",
					pathSlice[versionIndex], alias, path, alias, applicableAlias),
				SuggestedFixes: []analysis.SuggestedFix{{
					Message:   fmt.Sprintf("may replace %q with %q", alias, applicableAlias),
					TextEdits: findEdits(node, pass.TypesInfo.Uses, originalImportPath, alias, applicableAlias),
				}},
			})

			return
		}

		if err := checkAliasName(aliasSlice, pathSlice, pass); err != nil {
			applicableAlias := getAliasFix(pathSlice)
			pass.Report(analysis.Diagnostic{
				Pos:     node.Pos(),
				Message: fmt.Sprintf("%q may replace %q with %q", err.Error(), alias, applicableAlias),
				SuggestedFixes: []analysis.SuggestedFix{{
					Message:   fmt.Sprintf("may replace %q with %q", alias, applicableAlias),
					TextEdits: findEdits(node, pass.TypesInfo.Uses, originalImportPath, alias, applicableAlias),
				}},
			})

			return
		}
	})

	return nil, nil
}

// checkVersion checks that if package name starts with `v` it's included in alias name
func checkVersion(aliasLastWord string, pathSlice []string) bool {
	versionExists, versionPos := packageVersion(pathSlice)
	if !versionExists {
		return true
	}

	return aliasLastWord == pathSlice[versionPos]
}

// checkAliasName check consistency in alias name
func checkAliasName(aliasSlice []string, pathSlice []string, pass *analysis.Pass) error {
	lastUsedWordIndex := -1

	for _, name := range aliasSlice {
		// we don't check version rule here
		if strings.HasPrefix(name, "v") || name == "" {
			continue
		}
		usedWordIndex := searchString(pathSlice, name)

		if usedWordIndex == len(pathSlice) {
			return fmt.Errorf("alias %q does not contain any words from import path %q", strings.Join(aliasSlice, "_"), strings.Join(pathSlice, "/"))
		}

		if usedWordIndex <= lastUsedWordIndex {
			return fmt.Errorf("alias %q does not match word order from import path %q", strings.Join(aliasSlice, "_"), strings.Join(pathSlice, "/"))
		}

		lastUsedWordIndex = usedWordIndex
	}

	if lastUsedWordIndex == -1 {
		return fmt.Errorf("alias %q uses words that are not in path %q", strings.Join(aliasSlice, "_"), strings.Join(pathSlice, "/"))
	}

	return nil
}

func getAliasFix(pathSlice []string) string {
	versionExists, versionPos := packageVersion(pathSlice)

	if !versionExists {
		return pathSlice[len(pathSlice)-1]
	}

	if versionPos == len(pathSlice)-1 {
		applicableAlias := pathSlice[len(pathSlice)-2] + "_" + pathSlice[versionPos]
		return applicableAlias
	}

	applicableAlias := pathSlice[len(pathSlice)-1] + "_" + pathSlice[versionPos]

	return applicableAlias
}

// packageVersion returns if some version specification exists in import path and it's position
func packageVersion(pathSlice []string) (bool, int) {
	for pos, value := range pathSlice {
		r, _ := regexp.Compile("^v[0-9]+$")
		if r.MatchString(value) {
			return true, pos
		}
	}

	return false, 0
}

func searchString(slice []string, word string) int {
	for pos, value := range slice {
		r, _ := regexp.Compile("^" + word + "(s)?$")
		if r.MatchString(value) {
			return pos
		}
	}

	return len(slice)
}

func findEdits(node ast.Node, uses map[*ast.Ident]types.Object, importPath, original, required string) []analysis.TextEdit {
	// Edit the actual import line.
	result := []analysis.TextEdit{{
		Pos:     node.Pos(),
		End:     node.End(),
		NewText: []byte(required + " " + strconv.Quote(importPath)),
	}}

	// Edit all the uses of the alias in the code.
	for use, pkg := range uses {
		pkgName, ok := pkg.(*types.PkgName)
		if !ok {
			// skip identifiers that aren't pointing at a PkgName.
			continue
		}

		if pkgName.Pos() != node.Pos() {
			// skip identifiers pointing to a different import statement.
			continue
		}
		if original == pkgName.Name() {
			result = append(result, analysis.TextEdit{
				Pos:     use.Pos(),
				End:     use.End(),
				NewText: []byte(required),
			})
		}
	}
	return result
}
