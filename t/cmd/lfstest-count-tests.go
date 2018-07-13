package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	countFile = "test_count"
	lockFile  = "test_count.lock"

	errCouldNotAcquire = fmt.Errorf("could not acquire lock, dying")
)

type countFn func(int) (int, error)

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr,
			"usage: %s [increment|decrement]\n", os.Args[0])
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := acquire(ctx); err != nil {
		fatal(err)
	}
	defer release()

	if len(os.Args) == 1 {
		callWithCount(func(n int) (int, error) {
			fmt.Fprintf(os.Stdout, "%d\n", n)
			return n, nil
		})
		return
	}

	var err error

	switch strings.ToLower(os.Args[1]) {
	case "increment":
		err = callWithCount(func(n int) (int, error) {
			if n > 0 {
				return n + 1, nil
			}

			log, err := os.Create(fmt.Sprintf(
				"%s/gitserver.log", os.Getenv("LFSTEST_DIR")))
			if err != nil {
				return n, err
			}

			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("lfstest-gitserver.exe")
			} else {
				cmd = exec.Command("lfstest-gitserver")
			}
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("LFSTEST_URL=%s", os.Getenv("LFSTEST_URL")),
				fmt.Sprintf("LFSTEST_SSL_URL=%s", os.Getenv("LFSTEST_SSL_URL")),
				fmt.Sprintf("LFSTEST_CLIENT_CERT_URL=%s", os.Getenv("LFSTEST_CLIENT_CERT_URL")),
				fmt.Sprintf("LFSTEST_DIR=%s", os.Getenv("LFSTEST_DIR")),
				fmt.Sprintf("LFSTEST_CERT=%s", os.Getenv("LFSTEST_CERT")),
				fmt.Sprintf("LFSTEST_CLIENT_CERT=%s", os.Getenv("LFSTEST_CLIENT_CERT")),
				fmt.Sprintf("LFSTEST_CLIENT_KEY=%s", os.Getenv("LFSTEST_CLIENT_KEY")),
			)
			cmd.Stdout = log

			if err := cmd.Start(); err != nil {
				return n, err
			}
			return n + 1, nil
		})
	case "decrement":
		err = callWithCount(func(n int) (int, error) {
			if n > 1 {
				return n - 1, nil
			}

			url, err := ioutil.ReadFile(os.Getenv("LFS_URL_FILE"))
			if err == nil {
				_, err = http.Post(string(url)+"/shutdown",
					"application/text",
					strings.NewReader(time.Now().String()))
			}

			return 0, nil
		})
	}

	if err != nil {
		fatal(err)
	}
}

func acquire(ctx context.Context) error {
	path, err := path(lockFile)
	if err != nil {
		return err
	}

	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			_, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0666)
			if err == nil || !alreadyExists(err) {
				return err
			}
		case <-ctx.Done():
			return errCouldNotAcquire
		}
	}
}

func release() error {
	path, err := path(lockFile)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func callWithCount(fn countFn) error {
	path, err := path(countFile)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var n int = 0
	if len(contents) != 0 {
		n, err = strconv.Atoi(string(contents))
		if err != nil {
			return err
		}
	}

	after, err := fn(n)
	if err != nil {
		return err
	}

	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f, "%d", after); err != nil {
		return err
	}
	return nil
}

func path(s string) (string, error) {
	p := filepath.Join(filepath.Dir(os.Getenv("LFSTEST_DIR")), s)
	if err := os.MkdirAll(filepath.Dir(p), 0666); err != nil {
		return "", err
	}
	return p, nil
}

func alreadyExists(err error) bool {
	if err, ok := err.(*os.PathError); ok && err != nil {
		return err.Err.Error() == "file exists"
	}
	return false
}

func fatal(err error) {
	if err == nil {
		return
	}
	if err := release(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: while dying, got: %s\n", err)
	}
	fmt.Fprintf(os.Stderr, "fatal: %s\n", err)
	os.Exit(1)
}
