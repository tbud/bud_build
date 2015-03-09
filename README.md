# bud
**The golang build system use golang**

## Sample `build.bud`

This file is just a quick sample to give you a taste of what bud does.

```golang
#!/usr/bin/env bud

# config asset plugin
TaskConfig("bud.asset", Config{
		"patterns": []string{"bud.conf"},
		"output":   "context/assets.go",
		"package": "context",
})

UseTasks("bud")

Task("build", Tasks("asset"), func() error {
	return Exec("go", "get", "github.com/tbud/bud")
})

Task("default", func() error {
	Watch(Patterns("**/*.go"), Tasks("build"), func(events []Event) error {
		Printf("%v\n", events)
		return nil
	})
	return nil
})
```

The first line `#!/usr/bin/env bud` is option. You can use it, if you want use `./build.bud` to run bud command.

All tasks have a group name and a task name. You can run task with group name and task name, for example:
`bud bud.clean` will run bud.clean task to clean bud script run temp dirs.

`UseTasks` will use task name to create link to real task, then you can direct run task only use task name, for example:
In `build.bud` file add `UseTasks("bud")`, then you can run `bud clean` directly.

## Supported tasks
bud.asset - Package file into bin.

bud.clean - Clean bud script run temp dir.
