package cmd

import "github.com/MakeNowJust/heredoc"

var envHelp = map[string]string{
	"short": "List of supported environment variables",
	"long": heredoc.Doc(`
			RAYSTACK_CONFIG_DIR: the directory where frontier will store configuration files. Default:
			"$XDG_CONFIG_HOME/raystack" or "$HOME/.config/raystack".
			NO_COLOR: set to any value to avoid printing ANSI escape sequences for color output.
			CLICOLOR: set to "0" to disable printing ANSI colors in output.
		`),
}

var authHelp = map[string]string{
	"short": "Auth configs that need to be used with frontier",
	"long": heredoc.Doc(`
			Send an additional flag header with "key:value" format.
			Example:
				frontier create user -f user.yaml -H X-Frontier-Email:user@raystack.org
		`),
}
