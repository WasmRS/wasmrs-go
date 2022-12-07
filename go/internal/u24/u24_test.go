package u24_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/nanobus/iota/go/internal/common"
	"github.com/nanobus/iota/go/internal/u24"
)

func BenchmarkNewUint24(b *testing.B) {
	n := uint32(RandIntn(u24.MaxUint24))
	for i := uint32(0); i < uint32(b.N); i++ {
		_ = u24.MustNewUint24(n)
	}
}

func BenchmarkUint24_AsInt(b *testing.B) {
	n := u24.MustNewUint24(uint32(RandIntn(u24.MaxUint24)))
	for i := 0; i < b.N; i++ {
		_ = n.AsInt()
	}
}

func TestMustNewUint24(t *testing.T) {
	func() {
		defer func() {
			e := recover()
			assert.True(t, u24.IsExceedMaximumUint24Error(e.(error)), "should failed")
		}()
		_ = u24.MustNewUint24(u24.MaxUint24 + 1)
	}()
}

func TestUint24(t *testing.T) {
	testSingle(t, 0)
	for range [1000000]struct{}{} {
		testSingle(t, RandIntn(u24.MaxUint24))
	}
	testSingle(t, u24.MaxUint24)

	// over maximum number
	_, err := u24.NewUint24(u24.MaxUint24 + 1)
	assert.Error(t, err, "over maximum number should failed")
}

func TestUint24_WriteTo(t *testing.T) {
	var b bytes.Buffer
	for _, n := range []int{0, 1, RandIntn(u24.MaxUint24), u24.MaxUint24} {
		b.Reset()
		v := u24.MustNewUint24(uint32(n))
		wrote, err := v.WriteTo(&b)
		assert.NoError(t, err, "write uint24 failed")
		assert.Equal(t, int64(3), wrote, "wrote bytes length should be 3")
		assert.Equal(t, n, u24.NewUint24Bytes(b.Bytes()).AsInt(), "bad uint24 result")
	}
}

func testSingle(t *testing.T, n int) {
	x := u24.MustNewUint24(uint32(n))
	assert.Equal(t, n, x.AsInt(), "bad new from int")
	y := u24.NewUint24Bytes(x.Bytes())
	assert.Equal(t, n, y.AsInt(), "bad new from bytes")
}
