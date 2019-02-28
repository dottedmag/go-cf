package cf

// Taken from go-osx-plist (see LICENSE), heavily adapted

import (
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCFData(t *testing.T) {
	f := func(data []byte) []byte { return data }
	g := func(data []byte) []byte {
		pool := &Pool{}
		defer pool.Release()

		cfData := pool.Data(data)
		require.NotZero(t, cfData)
		return cfData.Goize()
	}
	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestCFString(t *testing.T) {
	// because the generator for string produces invalid strings,
	// lets generate []runes instead and convert those to strings in the function
	f := func(runes []rune) string { return string(runes) }
	g := func(runes []rune) string {
		pool := &Pool{}
		defer pool.Release()

		cfStr, err := pool.String(string(runes))
		require.NoError(t, err)
		return cfStr.Goize()
	}
	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestCFNumber_Int64(t *testing.T) {
	f := func(i int64) int64 { return i }
	g := func(i int64) int64 {
		pool := &Pool{}
		defer pool.Release()

		cfNum := pool.Int64(i)
		return cfNum.GoizeInt64()
	}
	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestCFNumber_UInt32(t *testing.T) {
	f := func(i uint32) uint32 { return i }
	g := func(i uint32) uint32 {
		pool := &Pool{}
		defer pool.Release()

		cfNum := pool.Uint32(i)
		return cfNum.GoizeUint32()
	}
	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestCFNumber_Float64(t *testing.T) {
	f := func(f float64) float64 { return f }
	g := func(f float64) float64 {
		pool := &Pool{}
		defer pool.Release()

		cfNum := pool.Float64(f)
		require.NotZero(t, cfNum)
		return cfNum.GoizeFloat64()
	}
	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestCFBoolean(t *testing.T) {
	f := func(b bool) bool { return b }
	g := func(b bool) bool {
		pool := &Pool{}
		defer pool.Release()

		cfBool := pool.Bool(b)
		require.NotZero(t, cfBool)
		return cfBool.Goize()
	}
	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestCFDate(t *testing.T) {
	// We know the CFDate conversion explicitly truncates to milliseconds
	// because CFDates use floating point for representation.
	round := func(nano int64) int64 {
		return int64(time.Duration(nano) / time.Millisecond * time.Millisecond)
	}
	f := func(nano int64) time.Time { return time.Unix(0, round(nano)) }
	g := func(nano int64) time.Time {
		pool := &Pool{}
		defer pool.Release()

		ti := time.Unix(0, round(nano))
		cfDate := pool.Date(ti)
		return cfDate.Goize()
	}
	if err := quick.CheckEqual(f, g, nil); err != nil {
		t.Error(err)
	}
}

func TestArbitrary(t *testing.T) {
	// test arbitrary values of any plistable type
	f := func(arb Arbitrary) interface{} { a, _ := standardize(arb.Value); return a }
	g := func(arb Arbitrary) interface{} {
		p := &Pool{}
		defer p.Release()

		if cfObj, err := p.refObject(reflect.ValueOf(arb.Value)); err != nil {
			t.Error(err)
		} else {
			if val, err := cfObj.Goize(); err != nil {
				t.Error(err)
			} else {
				a, _ := standardize(val)
				return a
			}
		}
		return nil
	}
	if err := quick.CheckEqual(f, g, nil); err != nil {
		if e, ok := err.(quick.SetupError); ok {
			t.Error(e)
			return
		}
		input := err.(*quick.CheckEqualError).In[0].(Arbitrary).Value
		t.Logf("Input value type: %T", input)
		t.Error(err)
	}
}
