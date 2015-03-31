// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builtin

import (
	"os"
	"os/exec"
)

func Exec(name string, args ...string) (err error) {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

type Command struct {
	exec.Cmd
}

func Cmd(name string, args ...string) *Command {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin

	cmd.Env = os.Environ()
	return &Command{(*cmd)}
}

func (c *Command) WithEnv(env ...string) *Command {
	c.Env = append(c.Env, env...)
	return c
}

func (c *Command) Run() error {
	return c.Cmd.Run()
}
