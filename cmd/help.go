package cmd

import "github.com/MakeNowJust/heredoc"

var envHelp = map[string]string{
	"short": "Environment variables that can be used with shield",
	"long": heredoc.Doc(`
			ODPF_CONFIG_DIR: the directory where shield will store configuration files. Default:
			"$XDG_CONFIG_HOME/odpf" or "$HOME/.config/odpf".
			NO_COLOR: set to any value to avoid printing ANSI escape sequences for color output.
			CLICOLOR: set to "0" to disable printing ANSI colors in output.
		`),
}

var authHelp = map[string]string{
	"short": "Auth configs that need to be used with shield",
	"long": heredoc.Doc(`
			Send an additional flag header with "key:value" format.
			Example:
				shield create user -f user.yaml -H X-Shield-Email:user@odpf.io
		`),
}
