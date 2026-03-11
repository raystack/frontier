# Shell Autocompletion

### Bash auto-completion

The Frontier completion script for Bash can be generated with `frontier completion bash`. Sourcing this script in your shell enables the Frontier completion.

However, the completion script depends on bash-completion, which means that you have to install this software first (you can test if you have bash-completion already installed by running
`type _init_completion`).

:::note
There are two versions of bash-completion, v1 and v2. V1 is for Bash 3.2 (which is the default on macOS), and v2 is for Bash 4.1+. The Frontier completion script doesn't work correctly with bash-completion v1 and Bash 3.2. It requires bash-completion v2 and Bash 4.1+. **Thus, to be able to correctly use Frontier completion on macOS, you have to install and use Bash 4.1+** (instructions). The following instructions assume that you use Bash 4.1+ (that is, any Bash version of 4.1 or newer).
:::

You now have to ensure that the Frontier completion script gets sourced in all your shell sessions. There are multiple ways to achieve this:

- Source the completion script in your ~/.bash_profile file:

```
echo 'source <(./frontier completion bash)' >> ~/.bash_profile
```

- Add the completion script to the /usr/local/etc/bash_completion.d directory:

```
# To load completions for each session, execute once:
# Linux:
$ ./frontier completion bash > /etc/bash_completion.d/_frontier

# macOS:
$ ./frontier completion bash > /usr/local/etc/bash_completion.d/_frontier
```

- If you installed Frontier with Homebrew (as explained in [getting started](../installation.md#macos)), then the frontier completion script should already be in /usr/local/etc/bash_completion.d/\_frontier. In that case, you don't need to do anything.

> Note: The Homebrew installation of bash-completion v2 sources all the files in the BASH_COMPLETION_COMPAT_DIR directory, that's why the latter two methods work.

In any case, after reloading your shell, Frontier completion should be working.

### Zsh Auto-completion

The Frontier completion script for Zsh can be generated with the command `frontier completion zsh`. Sourcing the completion script in your shell enables frontier autocompletion.

- If shell completion is not already enabled in your environment, you will need to enable it. You can execute the following once:

> If you get an error like `complete:13: command not found: compdef`, then add the following to the beginning of your `~/.zshrc` file:

```
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
```

- To load completions for each session, execute once:

```
  $ frontier completion zsh > "${fpath[1]}/_frontier"
```

- Now start a new shell for this setup to take effect and execute the below command to do sourcing in all your shell session:

```
  $ source ~/.zshrc
```

After setup is completed

```
 # Run the following command in shell (bash/zsh)
 $ frontier [tab][tab]
```

Output contains the list of all the core-commands which can be auto-completed now:

```
A cloud native role-based authorization aware reverse-proxy service.

Usage
  frontier <command> <subcommand> [flags]

Core commands
  group           Manage groups
  namespace       Manage namespaces
  organization    Manage organizations
  permission      Manage permissions
  policy          Manage policies
  project         Manage projects
  role            Manage roles
  user            Manage users

Other commands
  completion      Generate shell completion scripts
  config          Manage client configurations
  help            Help about any command
  server          Server management
  version         Print version information

Help topics
  auth            Auth configs that need to be used with frontier
  environment     List of supported environment variables
  reference       Comprehensive reference of all commands

Flags
  --help   Show help for command

Learn more
  Use 'frontier <command> <subcommand> --help' for info about a command.
  Read the manual at https://raystack.github.io/frontier/

Feedback
  Open an issue here https://github.com/raystack/frontier/issues
```
