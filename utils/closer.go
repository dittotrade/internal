package utils

import "io"

// CloseOrErr is a utility which handles defer part of func.
// It closes opened object and assign the result of operation to err which usually is the result of function
// assignment to err is completed only if Close returned and error and err is nil
// otherwise the result of close is ignored
func CloseOrErr(cl io.Closer, err *error) {
	e := cl.Close()
	if *err != nil {
		return
	}
	*err = e
}
