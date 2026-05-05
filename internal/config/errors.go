package config

import "fmt"

type ParseError struct {
    Line int
    Raw  string // original line text
    Msg  string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("ssh_config:%d: %s (raw: %q)", e.Line, e.Msg, e.Raw)
}