# Non-Interactive Mode

TicketFlow now supports non-interactive mode for CI/CD environments and automated scripts where user input is not available.

## How It Works

TicketFlow automatically detects non-interactive environments and uses sensible defaults instead of prompting for user input.

### Detection Methods

The following conditions trigger non-interactive mode:

1. **CI Environment Variables**: When common CI environment variables are set:
   - `CI`
   - `CONTINUOUS_INTEGRATION`
   - `GITHUB_ACTIONS`
   - `GITLAB_CI`
   - `CIRCLECI`
   - `JENKINS_URL`

2. **Explicit Control**: Set `TICKETFLOW_NON_INTERACTIVE=true`

3. **No TTY**: When stdin is not a terminal (e.g., when piping input)

### Behavior in Non-Interactive Mode

When non-interactive mode is detected:

1. **Prompts Use Defaults**: All prompts automatically use their default option
2. **No Default Available**: If a prompt has no default option, an error is returned
3. **Informative Output**: The system prints which option was automatically selected

## Branch Divergence Handling

When starting a ticket with an existing branch that has diverged:

- **Interactive Mode**: User is prompted to choose:
  - Use existing branch as-is
  - Delete and recreate branch at current HEAD (default)
  - Cancel operation

- **Non-Interactive Mode**: Automatically recreates the branch at current HEAD (the default option)

## Examples

### Running in CI

```bash
# GitHub Actions automatically sets CI=true
ticketflow start my-ticket

# Output:
# ⚠️  Branch 'my-ticket' already exists but has diverged from 'main'
#    • 1 commits ahead
#    • 2 commits behind
#
# Non-interactive mode detected. Using default option: r
# Recreating branch 'my-ticket' at current HEAD...
```

### Explicit Non-Interactive Mode

```bash
# Force non-interactive mode
export TICKETFLOW_NON_INTERACTIVE=true
ticketflow cleanup

# Cleanup prompts will use defaults automatically
```

### Testing in CI

Integration tests now work correctly in CI environments without hanging on prompts:

```bash
# In CI environments like GitHub Actions
make test

# Tests that involve prompts will automatically use defaults
```

## Best Practices

1. **Set Sensible Defaults**: Ensure your prompts have reasonable default options for CI environments

2. **Test Both Modes**: Test your scripts in both interactive and non-interactive modes

3. **Handle Errors**: In non-interactive mode, operations may fail if no default is available, so handle errors appropriately

4. **Use Force Flags**: Some commands like `cleanup` have `--force` flags that skip prompts entirely:
   ```bash
   ticketflow cleanup --force my-ticket
   ```

## Environment Variables

- `TICKETFLOW_NON_INTERACTIVE`: Set to `"true"` to force non-interactive mode
- Standard CI variables are automatically detected

## Troubleshooting

If you're experiencing issues with prompts in CI:

1. Check if CI environment variables are set: `env | grep CI`
2. Explicitly set `TICKETFLOW_NON_INTERACTIVE=true`
3. Use force flags where available to skip prompts entirely