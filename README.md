# AI Shell

## Usage with `cli2ssh`

```bash
cli2ssh -c "./aish" -e "SHELL_USERNAME={{ .User }}" -e "SHELL_COMMAND={{ .Command }}" -e "LOG_FILE={{ .RemoteAddr }}.log"
```
