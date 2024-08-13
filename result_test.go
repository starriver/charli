package charli_test

import (
	"testing"

	"github.com/starriver/charli"
)

var blankResult = charli.Result{
	App: &charli.App{},
}

func TestErrorString(t *testing.T) {
	r := blankResult
	r.ErrorString("test")

	if len(r.Errs) != 1 {
		t.Error("no error in Result")
	} else if s := r.Errs[0].Error(); s != "test" {
		t.Errorf("got '%s', want 'test'", s)
	}
}

func TestErrorf(t *testing.T) {
	r := blankResult
	r.Errorf("%s %d", "test", 123)

	if len(r.Errs) != 1 {
		t.Error("no error in Result")
	} else if s := r.Errs[0].Error(); s != "test 123" {
		t.Errorf("got '%s', want 'test 123'", s)
	}
}

func TestRunCommand(t *testing.T) {
	r := blankResult

	v := false

	r.Command = &charli.Command{
		Run: func(r2 *charli.Result) {
			if r2 != &r {
				t.Error("should be same result passed to run")
			}
			v = true
		},
	}

	r.RunCommand()

	if !v {
		t.Error("run func was ineffective")
	}
}
