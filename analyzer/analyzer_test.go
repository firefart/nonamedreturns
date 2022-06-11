package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %s", err)
	}

	testdata := filepath.Join(filepath.Dir(wd), "testdata")
	analysistest.Run(t, testdata, Analyzer, "default-config")

	err = Analyzer.Flags.Set(FlagAllowErrorInDefer, "true")
	if err != nil {
		t.Fatalf("Failed to set flag: %s", err)
	}
	analysistest.Run(t, testdata, Analyzer, "allow-error-in-defer")
}
