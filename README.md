### sample config
``` toml
[cron]
log_level = 'info'
[[cron.tasks]]
schedule = '@every 1s'
init = """
echo init >> test.txt
"""
cmd = """
echo hello >> test.txt
echo world >> test.txt
echo --- >> test.txt
"""
```