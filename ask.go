package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Question kinds.
const (
	kindNormal = iota
	kindConfirm
	kindChoice
)

// Question represents an interactive question.
type Question struct {
	question            string
	attempts            *int
	hidden              bool
	autocompleterValues []string
	validator           func(string) (interface{}, error)
	defaultVal          interface{}
	normalizer          func(string) string
	kind                int
}

// NewQuestion creates a normal question.
func NewQuestion(question string, defaultVal interface{}) *Question {
	return &Question{question: question, defaultVal: defaultVal, kind: kindNormal}
}

// GetQuestion returns the question text.
func (q *Question) GetQuestion() string { return q.question }

// GetDefault returns the default answer.
func (q *Question) GetDefault() interface{} { return q.defaultVal }

// IsHidden reports whether the answer is hidden.
func (q *Question) IsHidden() bool { return q.hidden }

// SetHidden enables hidden input.
func (q *Question) SetHidden(hidden bool) *Question {
	if hidden && len(q.autocompleterValues) > 0 {
		panic("a hidden question cannot use the autocompleter")
	}
	q.hidden = hidden
	return q
}

// GetAutocompleterValues returns the autocompletion candidates.
func (q *Question) GetAutocompleterValues() []string { return q.autocompleterValues }

// SetAutocompleterValues sets the autocompletion candidates.
func (q *Question) SetAutocompleterValues(values []string) *Question {
	if q.hidden {
		panic("a hidden question cannot use the autocompleter")
	}
	q.autocompleterValues = values
	return q
}

// GetValidator returns the validator.
func (q *Question) GetValidator() func(string) (interface{}, error) { return q.validator }

// SetValidator sets the validator.
func (q *Question) SetValidator(validator func(string) (interface{}, error)) *Question {
	q.validator = validator
	return q
}

// GetMaxAttempts returns the maximum number of attempts.
func (q *Question) GetMaxAttempts() *int { return q.attempts }

// SetMaxAttempts sets the maximum number of attempts.
func (q *Question) SetMaxAttempts(attempts int) *Question {
	if attempts < 1 {
		panic("maximum number of attempts must be a positive value")
	}
	q.attempts = &attempts
	return q
}

// GetNormalizer returns the normalizer function.
func (q *Question) GetNormalizer() func(string) string { return q.normalizer }

// SetNormalizer sets the normalizer function.
func (q *Question) SetNormalizer(normalizer func(string) string) *Question {
	q.normalizer = normalizer
	return q
}

// Confirmation is a yes/no question.
type Confirmation struct {
	*Question
	trueAnswerRegex string
}

// NewConfirmation creates a confirmation question.
func NewConfirmation(question string, defaultVal bool, trueAnswerRegex string) *Confirmation {
	if trueAnswerRegex == "" {
		trueAnswerRegex = "^y"
	}
	q := &Confirmation{
		Question:        NewQuestion(question, defaultVal),
		trueAnswerRegex: trueAnswerRegex,
	}
	q.kind = kindConfirm
	q.SetNormalizer(q.defaultNormalizer())
	return q
}

func (c *Confirmation) defaultNormalizer() func(string) string {
	def := c.GetDefault().(bool)
	regex := c.trueAnswerRegex
	return func(answer string) string {
		matched := regexpMatch(regex, answer)
		var result bool
		if !def {
			result = answer != "" && matched
		} else {
			result = answer == "" || matched
		}
		if result {
			return "yes"
		}
		return "no"
	}
}

// Choice is a multiple-choice question.
type Choice struct {
	*Question
	choices      map[string]string
	multiselect  bool
	prompt       string
	errorMessage string
}

// NewChoice creates a multiple-choice question; choices is a key->label mapping.
func NewChoice(question string, choices map[string]string, defaultVal interface{}) *Choice {
	if choices == nil {
		choices = map[string]string{}
	}
	c := &Choice{
		Question:     NewQuestion(question, defaultVal),
		choices:      choices,
		prompt:       " > ",
		errorMessage: `Value "%s" is invalid`,
	}
	c.kind = kindChoice
	c.SetValidator(c.defaultValidator())
	c.SetAutocompleterValues(keys(choices))
	return c
}

// GetChoices returns the choices.
func (c *Choice) GetChoices() map[string]string { return c.choices }

// SetMultiselect toggles multi-select.
func (c *Choice) SetMultiselect(multiselect bool) *Choice {
	c.multiselect = multiselect
	c.SetValidator(c.defaultValidator())
	return c
}

// IsMultiselect reports whether multi-select is enabled.
func (c *Choice) IsMultiselect() bool { return c.multiselect }

// GetPrompt returns the prompt.
func (c *Choice) GetPrompt() string { return c.prompt }

// SetPrompt sets the prompt.
func (c *Choice) SetPrompt(prompt string) *Choice {
	c.prompt = prompt
	return c
}

// SetErrorMessage sets the validation error message.
func (c *Choice) SetErrorMessage(msg string) *Choice {
	c.errorMessage = msg
	c.SetValidator(c.defaultValidator())
	return c
}

func (c *Choice) defaultValidator() func(string) (interface{}, error) {
	choices := c.choices
	errMsg := c.errorMessage
	multiselect := c.multiselect
	return func(selected string) (interface{}, error) {
		var selectedChoices []string
		if multiselect {
			rawChoices := strings.Split(selected, ",")
			selectedChoices = make([]string, 0, len(rawChoices))
			for _, v := range rawChoices {
				v = strings.TrimSpace(v)
				if v == "" {
					return nil, fmt.Errorf(errMsg, selected)
				}
				selectedChoices = append(selectedChoices, v)
			}
		} else {
			selectedChoices = []string{strings.TrimSpace(selected)}
		}

		result := make([]string, 0)
		for _, value := range selectedChoices {
			found := ""
			for key, label := range choices {
				if key == value || label == value {
					if found != "" {
						return nil, fmt.Errorf("the provided answer is ambiguous: %s", value)
					}
					found = key
				}
			}
			if found == "" {
				return nil, fmt.Errorf(errMsg, value)
			}
			result = append(result, found)
		}

		if multiselect {
			return result, nil
		}
		return result[0], nil
	}
}

// Ask handles interactive questions.
type Ask struct {
	out    *Output
	raw    interface{}
	custom io.Reader     // Input source injected via SetReader
	reader *bufio.Reader // Persistent buffered reader (avoids losing data from per-read prefetching)
}

// NewAsk creates a questioner.
func NewAsk(out *Output, raw interface{}) *Ask {
	return &Ask{out: out, raw: raw}
}

// SetReader overrides the input source (default os.Stdin), for testing and embedding.
// Once injected, this Ask uses its own independent buffered reader.
func (a *Ask) SetReader(r io.Reader) *Ask {
	a.custom = r
	a.reader = bufio.NewReader(r)
	return a
}

func (a *Ask) question() *Question {
	switch v := a.raw.(type) {
	case *Question:
		return v
	case *Confirmation:
		return v.Question
	case *Choice:
		return v.Question
	}
	return nil
}

func (a *Ask) choice() *Choice {
	if c, ok := a.raw.(*Choice); ok {
		return c
	}
	return nil
}

// Run asks the question and returns the validated answer.
func (a *Ask) Run() (interface{}, error) {
	q := a.question()
	if q == nil {
		return nil, nil
	}
	if q.GetValidator() == nil {
		ans, err := a.doAsk()
		if err != nil {
			return nil, err
		}
		// Without a validator, still apply the normalizer (e.g. Confirmation's yes/no normalization).
		if n := q.GetNormalizer(); n != nil {
			return n(fmt.Sprint(ans)), nil
		}
		return ans, nil
	}
	return a.validateAttempts(func() (interface{}, error) { return a.doAsk() })
}

func (a *Ask) validateAttempts(interviewer func() (interface{}, error)) (interface{}, error) {
	q := a.question()
	var err error
	remaining := -1
	if attempts := q.GetMaxAttempts(); attempts != nil {
		remaining = *attempts
	}
	for remaining == -1 || remaining > 0 {
		if remaining > 0 {
			remaining--
		}
		if err != nil {
			a.out.Error(err.Error())
		}
		answer, askErr := interviewer()
		if askErr != nil {
			return nil, askErr
		}
		validator := q.GetValidator()
		if validator == nil {
			return answer, nil
		}
		value, verr := validator(fmt.Sprint(answer))
		if verr == nil {
			if normalizer := q.GetNormalizer(); normalizer != nil {
				return normalizer(fmt.Sprint(value)), nil
			}
			return value, nil
		}
		err = verr
	}
	return nil, err
}

// bufioReader returns the persistent buffered reader: first the independent reader
// injected via SetReader, otherwise the shared reader of Output (guaranteeing that
// consecutive questions on the same output do not lose data).
func (a *Ask) bufioReader() *bufio.Reader {
	if a.reader != nil {
		return a.reader
	}
	if a.out != nil {
		return a.out.bufioReader()
	}
	return bufio.NewReader(os.Stdin)
}

func (a *Ask) doAsk() (interface{}, error) {
	a.writePrompt()
	q := a.question()

	var ret string
	var err error
	if q.IsHidden() {
		ret, err = a.getHiddenResponse()
	} else {
		ret, err = a.readResponse()
	}
	if err != nil {
		return nil, err
	}
	if ret != "" {
		return ret, nil
	}
	return q.GetDefault(), nil
}

func (a *Ask) readResponse() (string, error) {
	line, err := a.bufioReader().ReadString('\n')
	if err != nil && line == "" {
		return "", errors.New("aborted")
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (a *Ask) getHiddenResponse() (string, error) {
	if hasStty() {
		sttyMode, _ := exec.Command("stty", "-g").Output()
		_ = exec.Command("stty", "-echo").Run()
		defer func() {
			_ = exec.Command("stty", strings.TrimSpace(string(sttyMode))).Run()
		}()
		line, err := a.bufioReader().ReadString('\n')
		fmt.Fprintln(os.Stderr)
		if err != nil && line == "" {
			return "", errors.New("aborted")
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
	line, err := a.bufioReader().ReadString('\n')
	if err != nil && line == "" {
		return "", errors.New("aborted")
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (a *Ask) writePrompt() {
	// Do not print the prompt when no Output is bound (e.g. only SetReader was used).
	if a.out == nil {
		return
	}

	q := a.question()
	text := q.GetQuestion()
	def := q.GetDefault()

	switch q.kind {
	case kindConfirm:
		yesNo := "yes"
		if defBool, ok := def.(bool); ok && !defBool {
			yesNo = "no"
		}
		text = fmt.Sprintf(" %s (yes/no) [%s]:", text, yesNo)
	case kindChoice:
		choice := a.choice()
		if choice.IsMultiselect() {
			parts := strings.Split(fmt.Sprint(def), ",")
			labels := make([]string, 0)
			for _, p := range parts {
				if l, ok := choice.choices[strings.TrimSpace(p)]; ok {
					labels = append(labels, l)
				}
			}
			text = fmt.Sprintf(" %s [%s]:", text, strings.Join(labels, ", "))
		} else {
			label := fmt.Sprint(def)
			if l, ok := choice.choices[fmt.Sprint(def)]; ok {
				label = l
			}
			text = fmt.Sprintf(" %s [%s]:", text, label)
		}
	default:
		if def == nil {
			text = fmt.Sprintf(" %s:", text)
		} else {
			text = fmt.Sprintf(" %s [%v]:", text, def)
		}
	}

	a.out.raw(text)

	if q.kind == kindChoice {
		choice := a.choice()
		width := 0
		for k := range choice.choices {
			if len(k) > width {
				width = len(k)
			}
		}
		for _, k := range sortedKeys(choice.choices) {
			a.out.rawf("  [%-*s] %s", width, k, choice.choices[k])
		}
	}

	prompt := " > "
	if choice := a.choice(); choice != nil {
		prompt = choice.GetPrompt()
	}
	fmt.Fprint(os.Stdout, prompt)
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	out := keys(m)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func regexpMatch(pattern, s string) bool {
	matched, err := regexp.MatchString(pattern, s)
	return err == nil && matched
}

func hasStty() bool {
	_, err := exec.LookPath("stty")
	return err == nil
}
