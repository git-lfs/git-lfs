package main

import (
	"context"
	"fmt"
	"io"
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
	// countFile is the path to a file (relative to the $LFSTEST_DIR) who's
	// contents is the number of actively-running integration tests.
	countFile = "test_count"
	// lockFile is the path to a file (relative to the $LFSTEST_DIR) who's
	// presence indicates that another invocation of the lfstest-count-tests
	// program is modifying the test_count.
	lockFile = "test_count.lock"

	// lockAcquireTimeout is the maximum amount of time that we will wait
	// for lockFile to become available (and thus the amount of time that we
	// will wait in order to acquire the lock).
	lockAcquireTimeout = 5 * time.Second

	// errCouldNotAcquire indicates that the program could not acquire the
	// lock needed to modify the test_count. It is a fatal error.
	errCouldNotAcquire = fmt.Errorf("could not acquire lock, dying")
	// errNegativeCount indicates that the count in test_count was negative,
	// which is unexpected and makes this script behave in an undefined
	// fashion
	errNegativeCount = fmt.Errorf("unexpected negative count")
)

// countFn is a type signature that all functions who wish to modify the
// test_count must inhabit.
//
// The first and only formal parameter is the current number of running tests
// found in test_count after acquiring the lock.
//
// The returned tuple indicates (1) the new number that should be written to
// test_count, and (2) if there was an error in computing that value. If err is
// non-nil, the program will exit and test_count will not be updated.
type countFn func(int) (int, error)

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr,
			"usage: %s [increment|decrement]\n", os.Args[0])
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), lockAcquireTimeout)
	defer cancel()

	if err := acquire(ctx); err != nil {
		fatal(err)
	}
	defer release()

	if len(os.Args) == 1 {
		// Calling with no arguments indicates that we simply want to
		// read the contents of test_count.
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
				// If n>1, it is therefore true that a
				// lfstest-gitserver invocation is already
				// running.
				//
				// Hence, let's do nothing here other than
				// increase the count.
				return n + 1, nil
			}

			// The lfstest-gitserver invocation (see: below) does
			// not itself create a gitserver.log in the appropriate
			// directory. Thus, let's create it ourselves instead.
			log, err := os.Create(fmt.Sprintf(
				"%s/gitserver.log", os.Getenv("LFSTEST_DIR")))
			if err != nil {
				return n, err
			}

			// The executable name depends on the X environment
			// variable, which is set in script/cibuild.
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("lfstest-gitserver.exe")
			} else {
				cmd = exec.Command("lfstest-gitserver")
			}

			// The following are ported from the old
			// test/testhelpers.sh, and comprise the requisite
			// environment variables needed to run
			// lfstest-gitserver.
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
			cmd.Stderr = log

			// Start performs a fork/execve, hence we can abandon
			// this process once it has started.
			if err := cmd.Start(); err != nil {
				return n, err
			}
			return 1, nil
		})
	case "decrement":
		err = callWithCount(func(n int) (int, error) {
			if n > 1 {
				// If there is at least two tests running, we
				// need not shutdown a lfstest-gitserver
				// instance.
				return n - 1, nil
			}

			// Otherwise, we need to POST to /shutdown, which will
			// cause the lfstest-gitserver to abort itself.
			url, err := os.ReadFile(os.Getenv("LFS_URL_FILE"))
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

var (
	// acquireTick is the constant time that one tick (i.e., one attempt at
	// acquiring the lock) should last.
	acquireTick = 10 * time.Millisecond
)

// acquire acquires the lock file necessary to perform updates to test_count,
// and returns an error if that lock cannot be acquired.
func acquire(ctx context.Context) error {
	if disabled() {
		return nil
	}

	path, err := path(lockFile)
	if err != nil {
		return err
	}

	tick := time.NewTicker(acquireTick)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// Try every tick of the above ticker before giving up
			// and trying again.
			_, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0666)
			if err == nil || !os.IsExist(err) {
				return err
			}
		case <-ctx.Done():
			// If the context.Context above has reached its
			// deadline, we must give up.
			return errCouldNotAcquire
		}
	}
}

// release releases the lock file so that another process can take over, or
// returns an error.
func release() error {
	if disabled() {
		return nil
	}

	path, err := path(lockFile)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// callWithCount calls the given countFn with the current count in test_count,
// and updates it with what the function returns.
//
// If the function produced an error, that will be returned instead.
func callWithCount(fn countFn) error {
	path, err := path(countFile)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	contents, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var n int = 0
	if len(contents) != 0 {
		n, err = strconv.Atoi(string(contents))
		if err != nil {
			return err
		}

		if n < 0 {
			return errNegativeCount
		}
	}

	after, err := fn(n)
	if err != nil {
		return err
	}

	// We want to write over the contents in the file, so "truncate" the
	// file to a length of 0, and then seek to the beginning of the file to
	// update the write head.
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

// path returns an absolute path corresponding to any given path relative to the
// 't' directory of the current checkout of Git LFS.
func path(s string) (string, error) {
	p := filepath.Join(filepath.Dir(os.Getenv("LFSTEST_DIR")), s)
	if err := os.MkdirAll(filepath.Dir(p), 0666); err != nil {
		return "", err
	}
	return p, nil
}

// disabled returns true if and only if the lock acquisition phase is disabled.
func disabled() bool {
	s := os.Getenv("GIT_LFS_LOCK_ACQUIRE_DISABLED")
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}

// fatal reports the given error (if non-nil), and then dies. If the error was
// nil, nothing happens.
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
