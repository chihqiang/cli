package cli

import (
	"strings"
	"testing"
)

func TestNewQuestionDefaults(t *testing.T) {
	q := NewQuestion("name?", "world")
	if q.GetQuestion() != "name?" || q.GetDefault() != "world" {
		t.Errorf("question = %q default = %v", q.GetQuestion(), q.GetDefault())
	}
	if q.IsHidden() {
		t.Error("should not be hidden by default")
	}
}

func TestQuestionAutocompleterValues(t *testing.T) {
	q := NewQuestion("q", nil)
	q.SetAutocompleterValues([]string{"a", "b"})
	got := q.GetAutocompleterValues()
	if len(got) != 2 || got[0] != "a" {
		t.Errorf("autocompleter = %v", got)
	}
}

func TestQuestionHiddenConflict(t *testing.T) {
	q := NewQuestion("q", nil)
	q.SetAutocompleterValues([]string{"a"})
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected panic when hiding autocompleter question")
			}
		}()
		q.SetHidden(true)
	}()
}

func TestQuestionSetAutocompleterAfterHidden(t *testing.T) {
	q := NewQuestion("q", nil)
	q.SetHidden(true)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected panic setting autocompleter on hidden question")
			}
		}()
		q.SetAutocompleterValues([]string{"a"})
	}()
}

func TestSetMaxAttemptsPanic(t *testing.T) {
	q := NewQuestion("q", nil)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected panic for invalid attempts")
			}
		}()
		q.SetMaxAttempts(0)
	}()
	q.SetMaxAttempts(3)
	if q.GetMaxAttempts() == nil || *q.GetMaxAttempts() != 3 {
		t.Error("max attempts not set")
	}
}

func TestQuestionValidatorAndNormalizer(t *testing.T) {
	q := NewQuestion("q", nil)
	q.SetValidator(func(s string) (interface{}, error) { return s, nil })
	q.SetNormalizer(func(s string) string { return strings.ToUpper(s) })
	if q.GetValidator() == nil || q.GetNormalizer() == nil {
		t.Error("validator/normalizer not set")
	}
}

func TestConfirmationDefaultYes(t *testing.T) {
	c := NewConfirmation("continue?", true, "")
	n := c.defaultNormalizer()
	if n("y") != "yes" || n("") != "yes" || n("n") != "no" {
		t.Errorf("normalizer results wrong: %q %q %q", n("y"), n(""), n("n"))
	}
}

func TestConfirmationDefaultNo(t *testing.T) {
	c := NewConfirmation("continue?", false, "")
	n := c.defaultNormalizer()
	if n("") != "no" || n("yes") != "yes" || n("y") != "yes" {
		t.Errorf("normalizer results wrong: %q %q %q", n(""), n("yes"), n("y"))
	}
}

func TestConfirmationCustomRegex(t *testing.T) {
	c := NewConfirmation("proceed?", true, "^y(es)?$")
	n := c.defaultNormalizer()
	if n("yes") != "yes" || n("y") != "yes" || n("nope") != "no" {
		t.Errorf("custom regex normalizer wrong: %q %q %q", n("yes"), n("y"), n("nope"))
	}
}

func TestChoiceValidatorByKey(t *testing.T) {
	c := NewChoice("pick", map[string]string{"go": "Go", "py": "Python"}, "go")
	v := c.defaultValidator()
	got, err := v("py")
	if err != nil || got != "py" {
		t.Errorf("validator = %v err=%v", got, err)
	}
}

func TestChoiceValidatorByLabel(t *testing.T) {
	c := NewChoice("pick", map[string]string{"go": "Go", "py": "Python"}, "go")
	v := c.defaultValidator()
	got, err := v("Python")
	if err != nil || got != "py" {
		t.Errorf("validator by label = %v err=%v", got, err)
	}
}

func TestChoiceValidatorInvalid(t *testing.T) {
	c := NewChoice("pick", map[string]string{"go": "Go"}, "go")
	v := c.defaultValidator()
	if _, err := v("rust"); err == nil {
		t.Error("expected error for invalid choice")
	}
}

func TestChoiceValidatorAmbiguous(t *testing.T) {
	c := NewChoice("pick", map[string]string{"a": "x", "b": "x"}, "a")
	v := c.defaultValidator()
	if _, err := v("x"); err == nil {
		t.Error("expected ambiguity error")
	}
}

func TestChoiceMultiselect(t *testing.T) {
	c := NewChoice("pick", map[string]string{"go": "Go", "py": "Python"}, "go")
	c.SetMultiselect(true)
	if !c.IsMultiselect() {
		t.Error("should be multiselect")
	}
	v := c.defaultValidator()
	got, err := v("go, py")
	if err != nil {
		t.Fatalf("multiselect: %v", err)
	}
	gotSlice, ok := got.([]string)
	if !ok || len(gotSlice) != 2 || gotSlice[0] != "go" || gotSlice[1] != "py" {
		t.Errorf("multiselect result = %v", got)
	}
}

func TestChoiceMultiselectEmptyItem(t *testing.T) {
	c := NewChoice("pick", map[string]string{"go": "Go"}, "go")
	c.SetMultiselect(true)
	v := c.defaultValidator()
	if _, err := v("go, "); err == nil {
		t.Error("expected error for empty multiselect item")
	}
}

func TestChoicePromptAndErrorMsg(t *testing.T) {
	c := NewChoice("pick", map[string]string{"go": "Go"}, "go")
	if c.GetPrompt() != " > " {
		t.Errorf("prompt = %q", c.GetPrompt())
	}
	c.SetPrompt("==> ")
	if c.GetPrompt() != "==> " {
		t.Errorf("prompt after set = %q", c.GetPrompt())
	}
	c.SetErrorMessage("bad %s")
	if c.errorMessage != "bad %s" {
		t.Errorf("errorMessage = %q", c.errorMessage)
	}
	if len(c.GetChoices()) != 1 {
		t.Errorf("choices = %v", c.GetChoices())
	}
}

func TestAskRunWithStdinMultiQuestions(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("Zhang\n25\npy\n")
	o, _, _ := newBufOutput()
	o.root = app

	a1 := NewAsk(o, NewQuestion("name", "w"))
	v1, err := a1.Run()
	if err != nil || v1 != "Zhang" {
		t.Fatalf("q1 = %v err=%v", v1, err)
	}
	a2 := NewAsk(o, NewQuestion("age", 18))
	v2, err := a2.Run()
	if err != nil || v2 != "25" {
		t.Fatalf("q2 = %v err=%v", v2, err)
	}
	a3 := NewAsk(o, NewChoice("lang", map[string]string{"go": "Go", "py": "Python"}, "go"))
	v3, err := a3.Run()
	if err != nil || v3 != "py" {
		t.Fatalf("q3 = %v err=%v", v3, err)
	}
}

func TestAskRunUsesSharedReader(t *testing.T) {
	// Consecutive questions on the same Output share the buffered reader, avoiding data loss from prefetching.
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("a\nb\n")
	o, _, _ := newBufOutput()
	o.root = app

	q1 := NewAsk(o, NewQuestion("q1", nil))
	if v, err := q1.Run(); err != nil || v != "a" {
		t.Fatalf("q1 = %v err=%v", v, err)
	}
	q2 := NewAsk(o, NewQuestion("q2", nil))
	if v, err := q2.Run(); err != nil || v != "b" {
		t.Fatalf("q2 = %v err=%v", v, err)
	}
}

func TestAskSetReader(t *testing.T) {
	a := NewAsk(nil, NewQuestion("q", nil))
	a.SetReader(strings.NewReader("custom\n"))
	v, err := a.Run()
	if err != nil || v != "custom" {
		t.Fatalf("v=%v err=%v", v, err)
	}
}

func TestAskRunWithValidatorRetry(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("no\nok\n")
	o, _, _ := newBufOutput()
	o.root = app

	q := NewQuestion("q", nil)
	attempts := 3
	q.attempts = &attempts
	q.SetValidator(func(s string) (interface{}, error) {
		if s != "ok" {
			return nil, &UsageError{Msg: "must be ok"}
		}
		return s, nil
	})
	v, err := NewAsk(o, q).Run()
	if err != nil || v != "ok" {
		t.Fatalf("v=%v err=%v", v, err)
	}
}

func TestAskRunExhaustsAttempts(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("bad\nbad\n")
	o, _, _ := newBufOutput()
	o.root = app

	q := NewQuestion("q", nil)
	attempts := 1
	q.attempts = &attempts
	q.SetValidator(func(s string) (interface{}, error) {
		return nil, &UsageError{Msg: "always bad"}
	})
	if _, err := NewAsk(o, q).Run(); err == nil {
		t.Error("expected error after exhausting attempts")
	}
}

func TestAskRunHidden(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("secret\n")
	o, _, _ := newBufOutput()
	o.root = app

	q := NewQuestion("pwd", nil)
	q.SetHidden(true)
	v, err := NewAsk(o, q).Run()
	if err != nil || v != "secret" {
		t.Fatalf("v=%v err=%v", v, err)
	}
}

func TestAskRunAbortedOnEOF(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("")
	o, _, _ := newBufOutput()
	o.root = app

	q := NewQuestion("q", nil)
	if _, err := NewAsk(o, q).Run(); err == nil {
		t.Error("expected aborted error on EOF")
	}
}
