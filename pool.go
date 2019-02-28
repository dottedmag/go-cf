package cf

// Taken from go-osx-plist (see LICENSE), heavily adapted

// #import <CoreFoundation/CoreFoundation.h>
//
// CFArrayRef gocf_CFArrayCreate(CFAllocatorRef allocator, const uintptr_t *values, CFIndex numValues,
//    const CFArrayCallBacks *callBacks)
// {
//     return CFArrayCreate(allocator, (const void **)values, numValues, callBacks);
// }
//
// CFDictionaryRef gocf_CFDictionaryCreate(CFAllocatorRef allocator, const uintptr_t *keys, const uintptr_t *values,
//    CFIndex numValues, const CFDictionaryKeyCallBacks *keyCallBacks, const CFDictionaryValueCallBacks *valueCallBacks)
// {
//     return CFDictionaryCreate(allocator, (const void **)keys, (const void **)values, numValues,
//         keyCallBacks, valueCallBacks);
// }
//
import "C"
import (
	"fmt"
	"math"
	"reflect"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

type Pool struct {
	objects []typeRef
}

func (p *Pool) Release() {
	for _, o := range p.objects {
		C.CFRelease(C.CFTypeRef(o))
	}
	p.objects = nil
}

func (p *Pool) autorelease(obj typeRef) {
	if obj != 0 {
		p.objects = append(p.objects, obj)
	}
}

func (p *Pool) String(s string) (stringRef, error) {
	var cfs C.CFStringRef
	if s == "" {
		cfs = C.CFStringCreateWithBytes(0, nil, C.CFIndex(0), C.kCFStringEncodingUTF8, 0)
	} else {
		d := []byte(s)
		cfs = C.CFStringCreateWithBytes(0, (*C.UInt8)(unsafe.Pointer(&d[0])), C.CFIndex(len(d)),
			C.kCFStringEncodingUTF8, 0)
	}
	if cfs == 0 {
		return 0, fmt.Errorf("failed to convert %q to CFString", s)
	}
	p.autorelease(typeRef(cfs))
	return stringRef(cfs), nil
}

func (p *Pool) Bool(b bool) boolRef {
	var val C.CFBooleanRef
	if b {
		val = C.kCFBooleanTrue
	} else {
		val = C.kCFBooleanFalse
	}
	return boolRef(val)
}

func (p *Pool) Float64(f float64) numberRef {
	var v C.CFNumberRef
	if math.IsInf(f, 1) {
		v = C.kCFNumberPositiveInfinity
	} else if math.IsInf(f, -1) {
		v = C.kCFNumberNegativeInfinity
	} else if math.IsNaN(f) {
		v = C.kCFNumberNaN
	} else {
		double := C.double(f)
		v = C.CFNumberCreate(0, C.kCFNumberDoubleType, unsafe.Pointer(&double))
		p.autorelease(typeRef(v))
	}
	return numberRef(v)
}

func (p *Pool) Float32(f float32) numberRef {
	var v C.CFNumberRef
	if math.IsInf(float64(f), 1) {
		v = C.kCFNumberPositiveInfinity
	} else if math.IsInf(float64(f), -1) {
		v = C.kCFNumberNegativeInfinity
	} else if math.IsNaN(float64(f)) {
		v = C.kCFNumberNaN
	} else {
		float := C.float(f)
		v = C.CFNumberCreate(0, C.kCFNumberFloatType, unsafe.Pointer(&float))
		p.autorelease(typeRef(v))
	}
	return numberRef(v)
}

func (p *Pool) Int64(i int64) numberRef {
	sint64 := C.SInt64(i)
	v := C.CFNumberCreate(0, C.kCFNumberSInt64Type, unsafe.Pointer(&sint64))
	p.autorelease(typeRef(v))
	return numberRef(v)
}

func (p *Pool) Uint32(u uint32) numberRef {
	return p.Int64(int64(u))
}

func (p *Pool) Data(data []byte) dataRef {
	if len(data) == 0 {
		return dataRef(C.CFDataCreate(0, nil, 0))
	}
	ptr := (*C.UInt8)(&data[0])
	return dataRef(C.CFDataCreate(0, ptr, C.CFIndex(len(data))))
}

func (p *Pool) Date(t time.Time) dateRef {
	// truncate to milliseconds, to get a more predictable conversion
	ms := int64(time.Duration(t.UnixNano()) / time.Millisecond * time.Millisecond)
	nano := C.double(ms) / C.double(time.Second)
	nano -= C.double(C.kCFAbsoluteTimeIntervalSince1970)
	return dateRef(C.CFDateCreate(0, C.CFAbsoluteTime(nano)))
}

func (p *Pool) Object(i interface{}) (typeRef, error) {
	return p.refObject(reflect.ValueOf(i))
}

func (p *Pool) Array(i interface{}) (arrayRef, error) {
	slice := reflect.ValueOf(i)
	if slice.Kind() != reflect.Slice && slice.Kind() != reflect.Array {
		return 0, fmt.Errorf("non-slice in Array")
	}
	if slice.Len() == 0 {
		return arrayRef(C.CFArrayCreate(0, nil, 0, nil)), nil
	}
	cplists := []C.uintptr_t{}
	for i := 0; i < slice.Len(); i++ {
		obj, err := p.refObject(slice.Index(i))
		if err != nil {
			return 0, errors.Wrap(err, "failed to create CFArray")
		}
		cplists = append(cplists, C.uintptr_t(obj))
	}
	callbacks := (*C.CFArrayCallBacks)(&C.kCFTypeArrayCallBacks)
	return arrayRef(C.gocf_CFArrayCreate(0, &cplists[0], C.CFIndex(len(cplists)), callbacks)), nil
}

func (p *Pool) Dictionary(m interface{}) (dictionaryRef, error) {
	map_ := reflect.ValueOf(m)
	if map_.Kind() != reflect.Map {
		return 0, fmt.Errorf("non-map in Dictionary")
	}
	if map_.Type().Key().Kind() != reflect.String {
		return 0, fmt.Errorf("non-string map keys in Dictionary")
	}
	ckeys := []C.uintptr_t{}
	cvalues := []C.uintptr_t{}

	for _, key := range map_.MapKeys() {
		cfkey, err := p.String(key.String())
		if err != nil {
			return 0, err
		}
		ckeys = append(ckeys, C.uintptr_t(cfkey))

		cfval, err := p.refObject(map_.MapIndex(key))
		if err != nil {
			return 0, err
		}
		cvalues = append(cvalues, C.uintptr_t(cfval))
	}

	var keyPtr, valuePtr *C.uintptr_t
	if len(ckeys) > 0 {
		keyPtr = &ckeys[0]
		valuePtr = &cvalues[0]
	}

	keyCallbacks := (*C.CFDictionaryKeyCallBacks)(&C.kCFTypeDictionaryKeyCallBacks)
	valueCallbacks := (*C.CFDictionaryValueCallBacks)(&C.kCFTypeDictionaryValueCallBacks)
	return dictionaryRef(C.gocf_CFDictionaryCreate(0, keyPtr, valuePtr, C.CFIndex(len(ckeys)),
		keyCallbacks, valueCallbacks)), nil
}

func (p *Pool) refObject(v reflect.Value) (typeRef, error) {
	if !v.IsValid() {
		return 0, nil
	}
	switch v.Kind() {
	case reflect.Bool:
		return typeRef(p.Bool(v.Bool())), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return typeRef(p.Int64(v.Int())), nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return typeRef(p.Uint32(uint32(v.Uint()))), nil
	case reflect.Uint, reflect.Uintptr:
		// don't try and convert if uint/uintptr is 64-bits
		if v.Type().Bits() < 64 {
			return typeRef(p.Uint32(uint32(v.Uint()))), nil
		}
	case reflect.Float32:
		return typeRef(p.Float32(float32(v.Float()))), nil
	case reflect.Float64:
		return typeRef(p.Float64(v.Float())), nil
	case reflect.String:
		s, err := p.String(v.String())
		return typeRef(s), err
	case reflect.Struct:
		// only struct type we support is time.Time
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return typeRef(p.Date(v.Interface().(time.Time))), nil
		}
	case reflect.Array, reflect.Slice:
		// check for []byte first (byte is uint8)
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return typeRef(p.Data(v.Interface().([]byte))), nil
		}
		ary, err := p.Array(v.Interface())
		return typeRef(ary), err
	case reflect.Map:
		dict, err := p.Dictionary(v.Interface())
		return typeRef(dict), err
	case reflect.Interface:
		if v.IsNil() {
			return 0, &UnsupportedValueError{v, "nil interface"}
		}
		return p.refObject(v.Elem())
	}
	return 0, &UnsupportedTypeError{v.Type()}
}
