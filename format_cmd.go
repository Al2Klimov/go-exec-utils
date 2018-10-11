package go_exec_utils

import (
	"bytes"
)

var singleQuote = []byte("'")
var quotedSingleQuote = []byte(`'"'"'`)
var space = []byte(" ")

func FormatCmd(exe string, args []string, env map[string]string) string {
	res := make([][]byte, len(env)+1+len(args))
	i := 0

	for key, val := range env {
		res[i] = append(append([]byte(key), '='), quote4shell(val)...)
		i++
	}

	res[i] = quote4shell(exe)
	i++

	for _, arg := range args {
		res[i] = quote4shell(arg)
		i++
	}

	return string(bytes.Join(res, space))
}

func quote4shell(raw string) (quoted []byte) {
	replaced := bytes.Replace([]byte(raw), singleQuote, quotedSingleQuote, -1)
	quoted = make([]byte, len(replaced)+2)

	quoted[0] = singleQuote[0]
	copy(quoted[1:], replaced)
	quoted[len(quoted)-1] = singleQuote[0]

	return
}
