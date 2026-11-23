package gcfg

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-git/gcfg/v2/scanner"
	"github.com/go-git/gcfg/v2/token"
)

var unescape = map[rune]rune{'\\': '\\', '"': '"', 'n': '\n', 't': '\t', 'b': '\b'}

// no error: invalid literals should be caught by scanner
func unquote(s string) (string, error) {
	u, q, esc := make([]rune, 0, len(s)), false, false
	for _, c := range s {
		if esc {
			uc, ok := unescape[c]
			switch {
			case ok:
				u = append(u, uc)
				fallthrough
			case c == '\n':
				// A backslash at the end of a line continues the value
				// on the next one, and both are dropped, as git does.
				esc = false
				continue
			}
			return "", ErrMissingEscapeSequence
		}
		switch c {
		case '"':
			q = !q
		case '\\':
			esc = true
		case ' ', '\t', '\r':
			// Git drops unquoted whitespace while the value is still
			// empty. A quote does not count as content.
			if q || len(u) > 0 {
				u = append(u, c)
			}
		default:
			u = append(u, c)
		}
	}
	if q {
		return "", ErrMissingEndQuote
	}
	if esc {
		return "", ErrMissingEscapeSequence
	}
	return string(u), nil
}

func read(callback func(Event) error,
	fset *token.FileSet, file *token.File, src []byte) error {
	//
	var s scanner.Scanner
	var errs scanner.ErrorList
	s.Init(file, src, func(p token.Position, m string) { errs.Add(p, m) }, 0)
	sect, sectsub, hasSub := "", "", false
	var hdrPos token.Position
	pos, tok, lit, err := s.Scan()
	errfn := func(msg string) error {
		return fmt.Errorf("%s: %s", fset.Position(pos), msg)
	}
	if err != nil {
		return err
	}
	var accErr error
	for {
		if errs.Len() > 0 {
			if err, fatal := joinNonFatal(accErr, errs.Err()); fatal {
				return err
			}
		}
		switch tok {
		case token.EOF:
			return nil
		case token.EOL, token.COMMENT:
			pos, tok, lit, err = s.Scan()
			if err != nil {
				return err
			}
		case token.LBRACK:
			hdrPos = fset.Position(pos)
			pos, tok, lit, err = s.Scan()
			if err != nil {
				return err
			}
			if errs.Len() > 0 {
				if err, fatal := joinNonFatal(accErr, errs.Err()); fatal {
					return err
				}
			}
			if tok != token.IDENT {
				if err, fatal := joinNonFatal(accErr, errfn("expected section name")); fatal {
					return err
				}
			}
			// The scanner only returns a dot in the name of a legacy
			// [section.subsection] header. Git reads it as the section up
			// to the first dot and the subsection after it, both
			// lowercased.
			sect, sectsub, hasSub = lit, "", false
			if tok == token.IDENT {
				if name, sub, ok := strings.Cut(lit, "."); ok {
					sect, sectsub, hasSub = strings.ToLower(name), strings.ToLower(sub), true
				}
			}
			pos, tok, lit, err = s.Scan()
			if err != nil {
				return err
			}
			if errs.Len() > 0 {
				if err, fatal := joinNonFatal(accErr, errs.Err()); fatal {
					return err
				}
			}
			if tok == token.STRING {
				ss, err := unquote(lit)
				if err != nil {
					return err
				}

				sectsub, hasSub = ss, true
				pos, tok, lit, err = s.Scan()
				if err != nil {
					return err
				}
				if errs.Len() > 0 {
					if err, fatal := joinNonFatal(accErr, errs.Err()); fatal {
						return err
					}
				}
			}
			if tok != token.RBRACK {
				if err, fatal := joinNonFatal(accErr, errfn("expected right bracket")); fatal {
					return err
				}
			}
			pos, tok, lit, err = s.Scan()
			if err != nil {
				return err
			}
			if tok != token.EOL && tok != token.EOF && tok != token.COMMENT {
				if err, fatal := joinNonFatal(accErr, errfn("expected EOL, EOF, or comment")); fatal {
					return err
				}
			}
			// If a section/subsection header was found, ensure a
			// container object is created, even if there are no
			// variables further down.
			err := callback(Event{Section: sect, Subsection: sectsub, HasSubsection: hasSub,
				BlankValue: true, Position: hdrPos})
			if err != nil {
				return err
			}
		case token.IDENT:
			keyPos := fset.Position(pos)
			if sect == "" {
				if err, fatal := joinNonFatal(accErr, errfn("expected section header")); fatal {
					return err
				}
			}
			n := lit
			pos, tok, lit, err = s.Scan()
			if err != nil {
				return err
			}
			if errs.Len() > 0 {
				return errs.Err()
			}
			blank, v := tok == token.EOF || tok == token.EOL || tok == token.COMMENT, ""
			if !blank {
				if tok != token.ASSIGN {
					if err, fatal := joinNonFatal(accErr, errfn("expected '='")); fatal {
						return err
					}
				}
				pos, tok, lit, err = s.Scan()
				if err != nil {
					return err
				}
				if errs.Len() > 0 {
					if err, fatal := joinNonFatal(accErr, errs.Err()); fatal {
						return err
					}
				}
				if tok != token.STRING {
					if err, fatal := joinNonFatal(accErr, errfn("expected value")); fatal {
						return err
					}
				}
				unq, err := unquote(lit)
				if err != nil {
					return err
				}

				v = unq
				pos, tok, lit, err = s.Scan()
				if err != nil {
					return err
				}
				if errs.Len() > 0 {
					if err, fatal := joinNonFatal(accErr, errs.Err()); fatal {
						return err
					}
				}
				if tok != token.EOL && tok != token.EOF && tok != token.COMMENT {
					if err, fatal := joinNonFatal(accErr, errfn("expected EOL, EOF, or comment")); fatal {
						return err
					}
				}
			}
			err := callback(Event{Section: sect, Subsection: sectsub, HasSubsection: hasSub,
				Key: n, Value: v, BlankValue: blank, Position: keyPos})
			if err != nil {
				return err
			}
		default:
			if sect == "" {
				if err, fatal := joinNonFatal(accErr, errfn("expected section header")); fatal {
					return err
				}
			}
			if err, fatal := joinNonFatal(accErr, errfn("expected section header or variable declaration")); fatal {
				return err
			}
		}
	}
}

func readInto(config interface{}, fset *token.FileSet, file *token.File,
	src []byte) error {
	//
	firstPassCallback := func(e Event) error {
		return set(config, e.Section, e.Subsection, e.Key, e.Value, e.BlankValue, false)
	}
	err := read(firstPassCallback, fset, file, src)
	if err != nil {
		return err
	}
	secondPassCallback := func(e Event) error {
		return set(config, e.Section, e.Subsection, e.Key, e.Value, e.BlankValue, true)
	}
	return read(secondPassCallback, fset, file, src)
}

// ReadWithCallback reads gcfg formatted data from reader and calls
// callback with each section and option found.
//
// Callback is called with section, subsection, option key, option value
// and blank value flag as arguments.
//
// When a section is found, callback is called with an empty subsection,
// option key and option value.
//
// When a subsection is found, callback is called with an empty option key and
// option value.
//
// A header [x ""] is reported like [x], with the same empty subsection. A
// legacy [A.B] header is reported as section "a" and subsection "b".
// ReadWithOptions tells [x] and [x ""] apart.
//
// If blank value flag is true, it means that the value was not set for an option
// (as opposed to set to empty string).
//
// If callback returns an error, ReadWithCallback terminates with an error too.
func ReadWithCallback(reader io.Reader, callback func(string, string, string, string, bool) error) error {
	return ReadWithOptions(reader, WithCallback(func(e Event) error {
		return callback(e.Section, e.Subsection, e.Key, e.Value, e.BlankValue)
	}))
}

// Event is one item the reader found: a section header, or an option line
// under one. ReadWithOptions passes it to the callback set by WithCallback.
//
// A header is reported once, with an empty Key and Value and BlankValue true,
// so that a section without options is still seen. Each option line that
// follows is reported with the section and subsection of its header.
type Event struct {
	// Section is the section name, lowercased for a legacy [A.B] header,
	// whose subsection is lowercased too.
	Section string
	// Subsection is the subsection name, empty for [x] and for [x ""].
	Subsection string
	// HasSubsection tells [x] from [x ""], which both give an empty
	// Subsection but which Git reads as x.key and x..key. It is true for a
	// header with a quoted subsection, including an empty one, and for a
	// legacy [section.subsection] header such as [a.b], and false for a
	// plain [x]. It describes the header being read, so it is also set on
	// the option lines that follow that header.
	HasSubsection bool
	// Key and Value are the option name and its value, both empty for a
	// header.
	Key, Value string
	// BlankValue is true when the option has no value (as opposed to an
	// empty string), and for a header.
	BlankValue bool
	// Position is where the item starts: the "[" of a header, or the key of
	// an option line. Its Line and Column start at 1, and Filename is empty.
	// A consumer can use it to point at the item in an error message.
	Position token.Position
}

// ReadOption configures ReadWithOptions.
type ReadOption func(*readOptions)

type readOptions struct {
	callback func(Event) error
}

// WithCallback sets the function ReadWithOptions calls with each Event. If it
// returns an error, reading stops and ReadWithOptions returns that error. A
// later WithCallback replaces an earlier one.
func WithCallback(callback func(Event) error) ReadOption {
	return func(o *readOptions) { o.callback = callback }
}

// ReadWithOptions reads gcfg formatted data from reader and reports each
// section header and option to the callback set by WithCallback. Without it
// the data is only parsed: no events are reported and the errors are the same
// as with a callback that ignores every event.
//
// It is ReadWithCallback with an Event in place of positional arguments, so
// that it can also report whether a header had a subsection. The Read*Into
// functions do not use that flag, so an empty subsection there still folds
// onto the section.
func ReadWithOptions(reader io.Reader, opts ...ReadOption) error {
	o := readOptions{callback: func(Event) error { return nil }}
	for _, opt := range opts {
		opt(&o)
	}

	src, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	file, err := fset.AddFile("", fset.Base(), len(src))
	if err != nil {
		return err
	}

	return read(o.callback, fset, file, src)
}

// ReadInto reads gcfg formatted data from reader and sets the values into the
// corresponding fields in config.
//
// Git files the keys under [sec ""] as sec..key, which a struct target cannot
// represent, so ReadInto reads them as sec.key, like those under [sec]. A
// caller that applies a policy keyed on a subsection must use
// ReadWithOptions instead.
func ReadInto(config interface{}, reader io.Reader) error {
	src, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := fset.AddFile("", fset.Base(), len(src))
	if err != nil {
		return err
	}
	return readInto(config, fset, file, src)
}

// ReadStringInto reads gcfg formatted data from str and sets the values into
// the corresponding fields in config.
//
// See ReadInto for how an empty subsection is read.
func ReadStringInto(config interface{}, str string) error {
	r := strings.NewReader(str)
	return ReadInto(config, r)
}

// ReadFileInto reads gcfg formatted data from the file filename and sets the
// values into the corresponding fields in config.
//
// See ReadInto for how an empty subsection is read.
func ReadFileInto(config interface{}, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	src, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := fset.AddFile(filename, fset.Base(), len(src))
	if err != nil {
		return err
	}
	return readInto(config, fset, file, src)
}
