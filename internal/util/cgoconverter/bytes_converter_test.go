package cgoconverter

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
)

func TestBytesConverter(t *testing.T) {
	data := []byte("bytes converter test\n")
	length := len(data)
	cbytes := copyToCBytes(data)

	lease, goBytes := UnsafeGoBytes(&cbytes, length)
	defer Release(lease)
	equalBytes(t, data, goBytes)

	// data = make([]byte, maxByteArrayLen)
	v := byte(0x57)
	length = maxByteArrayLen
	cbytes = mallocCBytes(v, maxByteArrayLen)

	lease, goBytes = UnsafeGoBytes(&cbytes, length)
	defer Release(lease)

	if !isAll(goBytes, v) {
		t.Errorf("incorrect value, all bytes should be %v", v)
	}

	// equalBytes(t, data, goBytes)
}

func TestConcurrentBytesConverter(t *testing.T) {
	concurrency := runtime.GOMAXPROCS(0)
	if concurrency <= 1 {
		concurrency = 4
	}

	length := maxByteArrayLen / concurrency

	errCh := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		go func(iter int) {
			v := byte(iter)
			cbytes := mallocCBytes(v, length)

			lease, goBytes := UnsafeGoBytes(&cbytes, length)
			defer Release(lease)

			if !isAll(goBytes, v) {
				errCh <- fmt.Errorf("iter %d: incorrect value, all bytes should be %v", iter, v)
			} else {
				errCh <- nil
			}
		}(i)
	}

	hasErr := false
	for i := 0; i < concurrency; i++ {
		err := <-errCh
		if err != nil {
			t.Logf("err=%+v", err)
			hasErr = true
		}
	}
	if hasErr {
		t.Fail()
	}
}

func equalBytes(t *testing.T, origin []byte, new []byte) {
	if len(origin) != len(new) {
		t.Errorf("len(new)=%d new=%+v", len(new), new)
	}

	if !bytes.Equal(origin, new) {
		t.Errorf("data is not consistent")
	}
}

func isAll(data []byte, v byte) bool {
	for _, b := range data {
		if b != v {
			return false
		}
	}

	return true
}
