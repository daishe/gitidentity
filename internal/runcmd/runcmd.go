package runcmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/daishe/gitidentity/internal/logging"
)

func CommandError(cmd string, out []byte, err error) error {
	if ee := (&exec.ExitError{}); errors.As(err, &ee) {
		err = fmt.Errorf("command %q returned non zero exit code %d", cmd, ee.ExitCode())
	} else {
		err = fmt.Errorf("command %q unexpected error: %w", cmd, err)
	}
	if len(out) != 0 {
		err = fmt.Errorf("%w, command output:\n%s", err, out)
	}
	return err
}

func CommandCombinedOutput(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	logging.Log.Printf("COMMAND: %s %s", cmd, strings.Join(args, " "))
	out, err := exec.CommandContext(ctx, cmd, args...).CombinedOutput()
	if ee := (&exec.ExitError{}); errors.As(err, &ee) {
		logging.Log.Printf("COMMAND FAILED (NON ZERO EXIT CODE): %s %s, err: %v, output: %q", cmd, strings.Join(args, " "), err, string(out))
		return out, err
	}
	if err != nil {
		logging.Log.Printf("COMMAND FAILED: %s %s, err: %v, output: %q", cmd, strings.Join(args, " "), err, string(out))
		return out, err
	}
	logging.Log.Printf("COMMAND RETURNED: %s %s, output: %q", cmd, strings.Join(args, " "), string(out))
	return out, nil
}

func CommandPipeOutputAndInput(ctx context.Context, cmd string, args ...string) error {
	logging.Log.Printf("COMMAND: %s %s", cmd, strings.Join(args, " "))
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if ee := (&exec.ExitError{}); errors.As(err, &ee) {
		os.Exit(ee.ExitCode())
	}
	if err != nil {
		logging.Log.Printf("COMMAND FAILED: %s %s, err: %v", cmd, strings.Join(args, " "), err)
		return err
	}
	logging.Log.Printf("COMMAND RETURNED: %s %s", cmd, strings.Join(args, " "))
	return nil
}
