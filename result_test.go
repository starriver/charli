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
