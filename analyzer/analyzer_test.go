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

	err = Analyzer.Flags.Set(FlagReportErrorInDefer, "true")
	if err != nil {
		t.Fatalf("Failed to set flag: %s", err)
	}
	analysistest.Run(t, testdata, Analyzer, "report-error-in-defer")

	if err := Analyzer.Flags.Set(FlagReportErrorInDefer, "false"); err != nil {
		t.Fatalf("Failed to reset flag: %s", err)
	}
	if err := Analyzer.Flags.Set(FlagAllowUnusedNamedReturns, "true"); err != nil {
		t.Fatalf("Failed to set flag: %s", err)
	}
	analysistest.Run(t, testdata, Analyzer, "allow-unused-named-returns")

	// the flags are process-global state on the Analyzer, so reset them for
	// anything that runs after this test
	if err := Analyzer.Flags.Set(FlagAllowUnusedNamedReturns, "false"); err != nil {
		t.Fatalf("Failed to reset flag: %s", err)
	}
}
