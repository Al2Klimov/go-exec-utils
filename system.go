package go_exec_utils

import (
	"bytes"
	"context"
	"golang.org/x/sync/semaphore"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

type exePathInfo struct {
	path  string
	mutex sync.RWMutex
}

var exePaths = map[string]*exePathInfo{}
var exePathsMutex sync.RWMutex
var execSemaphore = semaphore.NewWeighted(int64(runtime.NumCPU()) * 2)
var execSemaphoreContext = context.Background()

func System(exe string, args []string, env map[string]string, cwd string) (effCmd string, out []byte, err error) {
	exePathsMutex.RLock()
	pathInfo, hasPath := exePaths[exe]
	exePathsMutex.RUnlock()

	if !hasPath {
		exePathsMutex.Lock()

		if pathInfo, hasPath = exePaths[exe]; !hasPath {
			pathInfo = &exePathInfo{path: ""}
			exePaths[exe] = pathInfo
		}

		exePathsMutex.Unlock()
	}

	pathInfo.mutex.RLock()
	exePath := pathInfo.path
	pathInfo.mutex.RUnlock()

	if exePath == "" {
		pathInfo.mutex.Lock()

		if exePath = pathInfo.path; exePath == "" {
			path, errLP := exec.LookPath(exe)
			if errLP != nil {
				pathInfo.mutex.Unlock()
				return FormatCmd(exe, args, env), nil, errLP
			}

			exePath = path
			pathInfo.path = exePath
		}

		pathInfo.mutex.Unlock()
	}

	cmd := exec.Command(exePath, args...)
	outBuf := bytes.Buffer{}

	flatEnv := make([]string, len(env))
	i := 0

	for key, val := range env {
		flatEnv[i] = key + "=" + val
		i++
	}

	cmd.Env = flatEnv
	cmd.Dir = cwd
	cmd.Stdin = nil
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr

	effCmd = FormatCmd(exePath, args, env)

	execSemaphore.Acquire(execSemaphoreContext, 1)
	err = cmd.Run()
	execSemaphore.Release(1)

	out = outBuf.Bytes()

	return
}
