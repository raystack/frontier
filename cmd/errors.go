package cmd

import (
	"errors"

	"github.com/MakeNowJust/heredoc"
)

var (
	ErrClientConfigNotFound = errors.New(heredoc.Doc(`
		Frontier client config not found.

		Run "frontier config init" to initialize a new client config or
		Run "frontier help environment" for more information.
	`))
	ErrClientConfigHostNotFound = errors.New(heredoc.Doc(`
		Frontier client config "host" not found.

		Pass frontier server host with "--host" flag or 
		set host in frontier config.

		Run "frontier config <subcommand>" or
		"frontier help environment" for more information.
	`))
	ErrClientNotAuthorized = errors.New(heredoc.Doc(`
		Frontier auth error. Frontier requires an auth header.
		
		Run "frontier help auth" for more information.
	`))
)
