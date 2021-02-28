package messagefmt

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
			desc: "logrus",
			pkg:  "a",
		},
		{
			desc: "kingpin",
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
