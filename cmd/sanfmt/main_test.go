package main

import (
	"bytes"
	"os"
	"testing"
)

func captureOutput(t *testing.T, f func()) (stdout, stderr string) {
	t.Helper()

	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = wOut, wErr
	defer func() { os.Stdout, os.Stderr = oldOut, oldErr }()

	f()

	wOut.Close()
	wErr.Close()
	var outBuf, errBuf bytes.Buffer
	outBuf.ReadFrom(rOut)
	errBuf.ReadFrom(rErr)
	return outBuf.String(), errBuf.String()
}

func TestRunAlgebraic(t *testing.T) {
	var code int
	out, errOut := captureOutput(t, func() {
		code = run([]string{"1. e4", "N x e4 ch"}, false, true)
	})
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty", errOut)
	}
	want := "e4\nNxe4+\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}

func TestRunDescriptiveAlternatesSides(t *testing.T) {
	var code int
	out, _ := captureOutput(t, func() {
		code = run([]string{"P-K4", "P-K4", "N-KB3"}, true, true)
	})
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	want := "e4\ne5\nNf3\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}

func TestRunDescriptiveBlackFirst(t *testing.T) {
	var code int
	out, _ := captureOutput(t, func() {
		code = run([]string{"P-K4", "P-K4"}, true, false)
	})
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	want := "e5\ne4\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}

func TestRunReportsFailuresButContinues(t *testing.T) {
	var code int
	out, errOut := captureOutput(t, func() {
		code = run([]string{"e4", "not-a-move", "e5"}, false, true)
	})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if errOut == "" {
		t.Error("stderr = empty, want a parse error reported")
	}
	want := "e4\ne5\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}

func TestRunSideAlternatesAcrossFailures(t *testing.T) {
	var code int
	out, _ := captureOutput(t, func() {
		code = run([]string{"P-K4", "NxP", "P-K4"}, true, true)
	})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	want := "e4\ne5\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}
