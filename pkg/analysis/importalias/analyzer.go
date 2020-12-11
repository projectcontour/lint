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
	"regexp"
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
		aliasSlice := strings.Split(alias, "_")
		path := strings.ReplaceAll(importStmt.Path.Value, "\"", "")
		// replace all separators with `/` for normalization
		path = strings.ReplaceAll(path, "_", "/")
		path = strings.ReplaceAll(path, ".", "/")
		path = strings.ReplaceAll(path, "-", "")
		// omit the domain name in path
		pathSlice := strings.Split(path, "/")[1:]
		packageName := pathSlice[len(pathSlice)-1]

		if !checkVersion(aliasSlice[len(aliasSlice)-1], pathSlice) {
			applicableAlias := getAliasFix(pathSlice)
			pass.Report(
				analysis.Diagnostic{
					Pos:     node.Pos(),
					Message: fmt.Sprintf("version not specified in alias. path: %s alias: %s version %s", path, alias, packageName),
					SuggestedFixes: []analysis.SuggestedFix{
						{
							Message: fmt.Sprintf("should replace %q with %q", alias, applicableAlias),
							TextEdits: []analysis.TextEdit{
								{
									Pos:     importStmt.Pos(),
									End:     importStmt.Name.End(),
									NewText: []byte(applicableAlias),
								},
							},
						},
					},
				},
			)
			return
		}
		if ok, lintErrMsg := checkAliasName(aliasSlice, pathSlice, pass); !ok {
			applicableAlias := getAliasFix(pathSlice)
			pass.Report(
				analysis.Diagnostic{
					Pos:     node.Pos(),
					Message: fmt.Sprintf(lintErrMsg+" path: %s alias: %s", path, alias),
					SuggestedFixes: []analysis.SuggestedFix{
						{
							Message: fmt.Sprintf("should replace %q with %q", alias, applicableAlias),
							TextEdits: []analysis.TextEdit{
								{
									Pos:     importStmt.Pos(),
									End:     importStmt.Name.End(),
									NewText: []byte(applicableAlias),
								},
							},
						},
					},
				},
			)
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
func checkAliasName(aliasSlice []string, pathSlice []string, pass *analysis.Pass) (bool, string) {
	lastUsedWordIndex := -1
	for _, name := range aliasSlice {
		// we don't check version rule here
		if strings.HasPrefix(name, "v") || name == "" {
			continue
		}
		usedWordIndex := searchString(pathSlice, name)

		if usedWordIndex == len(pathSlice) {
			return false, "used words in alias most be present in path"
		}

		if usedWordIndex <= lastUsedWordIndex {
			return false, "order of words in alias should match words in path"
		}

		lastUsedWordIndex = usedWordIndex
	}

	if lastUsedWordIndex == -1 {
		return false, "at least one word from path must be present in alias"
	}

	return true, ""
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
		if strings.HasPrefix(value, "v") {
			return true, pos
		}
	}
	return false, 0
}

func searchString(slice []string, word string) int {
	for pos, value := range slice {
		r, _ := regexp.Compile(word + "(s)?")
		if r.MatchString(value) {
			return pos
		}
	}
	return len(slice)
}
