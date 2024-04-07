# AI Shell

## Usage with `cli2ssh`

```bash
cli2ssh -h 0.0.0.0 -c "./aish" -e "AISH_USERNAME={{ .User }}" -e "AISH_COMMAND={{ .Command }}" -e "LOG_FILE={{ .RemoteAddr }}.log"
```
