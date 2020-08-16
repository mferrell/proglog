// START: intro
package log_test

import (
	"io/ioutil"
	"os"
	"testing"

	api "github.com/mferrell/proglog/api/v1"
	"github.com/mferrell/proglog/internal/log"
	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, log *log.Log){
		"append and read a batch succeeds": testAppendRead,
		"offset out of range error":        testOutOfRangeErr,
		"init with existing segments":      testInitExisting,
	} {
		t.Run(scenario, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "store-test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			c := log.Config{}
			c.Segment.MaxStoreBytes = 32
			log, err := log.NewLog(dir, c)
			require.NoError(t, err)

			fn(t, log)
		})
	}
}

// END: intro

// START: tests
func testAppendRead(t *testing.T, log *log.Log) {
	append := &api.Record{
		Value: []byte("hello world"),
	}
	off, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, append, read)

}

func testOutOfRangeErr(t *testing.T, log *log.Log) {
	read, err := log.Read(1)
	require.Nil(t, read)
	apiErr := err.(api.ErrOffsetOutOfRange)
	require.Equal(t, uint64(1), apiErr.Offset)
}

func testInitExisting(t *testing.T, o *log.Log) {
	append := &api.Record{
		Value: []byte("hello world"),
	}
	for i := 0; i < 3; i++ {
		_, err := o.Append(append)
		require.NoError(t, err)
	}
	require.NoError(t, o.Close())

	n, err := log.NewLog(o.Dir, o.Config)
	require.NoError(t, err)
	off, err := n.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(3), off)
}

// END: tests
