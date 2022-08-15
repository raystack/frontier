package cmd

import (
	"errors"

	"github.com/MakeNowJust/heredoc"
)

var (
	ErrClientConfigNotFound = errors.New(heredoc.Doc(`
		Shield client config not found.

		Run "shield config init" or
		Run "shield help environment" for more information.
	`))
	ErrClientConfigHostNotFound = errors.New(heredoc.Doc(`
		Shield client config "host" not found.

		Pass shield server host with "--host" flag or 
		use shield config.

		Run "shield config <subcommand>" or
		"shield help environment" for more information.
	`))
	ErrClientNotAuthorized = errors.New(heredoc.Doc(`
		Shield auth error. Shield requires an auth header.
		
		Run "shield help auth" for more information.
	`))
)
