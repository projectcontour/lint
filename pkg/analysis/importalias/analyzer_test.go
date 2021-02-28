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
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	testCases := []struct {
		desc string
		pkg  string
	}{
		{
			desc: "Valid imports",
			pkg:  "a",
		},
		{
			desc: "Invalid imports",
			pkg:  "b",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(testdata, "src", test.pkg)

			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				cmd := exec.Command("go", "mod", "vendor")
				cmd.Dir = dir

				t.Cleanup(func() {
					_ = os.RemoveAll(filepath.Join(testdata, "src", test.pkg, "vendor"))
				})

				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatal(err, string(output))
				}
			}

			analysistest.RunWithSuggestedFixes(t, testdata, Analyzer, test.pkg)
		})
	}
}
