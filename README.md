### sample config
``` toml
[cron]
log_level = 'info'
[[cron.tasks]]
name = "cron-demo"
schedule = '@every 1s'
init = """
echo init >> test.txt
"""
cmd = """
echo hello >> test-cron.txt
echo world >> test-cron.txt
echo --- >> test-cron.txt
"""
[watcher]
log_level = 'info'
[[watcher.tasks]]
name = "watcher-demo"
path = './test.txt'
init = """
echo init >> test-watcher.txt
"""
cmd = """
echo hello >> test-watcher.txt
echo world >> test-watcher.txt
echo --- >> test-watcher.txt
"""
```