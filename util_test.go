package xxljob_test

import (
	"testing"
	"time"

	"github.com/hyperjiang/xxljob"
	"github.com/stretchr/testify/require"
)

func TestLocalIP(t *testing.T) {
	should := require.New(t)

	should.NotEmpty(xxljob.LocalIP())
	should.Contains(xxljob.LocalIP(), ".")
}

func TestReadableSize(t *testing.T) {
	should := require.New(t)

	should.Equal("1b", xxljob.ReadableSize(1))
	should.Equal("1.95kb", xxljob.ReadableSize(2000))
	should.Equal("10.74kb", xxljob.ReadableSize(11000))
	should.Equal("1.14mb", xxljob.ReadableSize(1200000))
	should.Equal("1.02gb", xxljob.ReadableSize(1100000000))
}

func TestTruncateDuration(t *testing.T) {
	should := require.New(t)

	should.Equal("2s", xxljob.TruncateDuration(time.Second*2).String())
	should.Equal("500ms", xxljob.TruncateDuration(time.Millisecond*500).String())
	should.Equal("10µs", xxljob.TruncateDuration(time.Microsecond*10).String())
	should.Equal("0s", xxljob.TruncateDuration(time.Nanosecond*100).String())
}
