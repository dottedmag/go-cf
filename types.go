package cf

// Taken from go-osx-plist (see LICENSE), heavily adapted

// #import <CoreFoundation/CoreFoundation.h>
// #import <ApplicationServices/ApplicationServices.h> // CGFloat
// #cgo LDFLAGS: -framework CoreFoundation
import "C"
import (
	"time"
	"unsafe"
)

type typeRef C.CFTypeRef

func Release(i interface{}) {
	o := i.(typeRef)
	if o != 0 {
		C.CFRelease(C.CFTypeRef(o))
	}
}

func (t typeRef) Goize() (interface{}, error) {
	if t == 0 {
		return nil, nil
	}
	typeId := C.CFGetTypeID(C.CFTypeRef(t))
	switch typeId {
	case C.CFStringGetTypeID():
		return stringRef(t).Goize(), nil
	case C.CFNumberGetTypeID():
		return numberRef(t).Goize(), nil
	case C.CFBooleanGetTypeID():
		return boolRef(t).Goize(), nil
	case C.CFDataGetTypeID():
		return dataRef(t).Goize(), nil
	case C.CFDateGetTypeID():
		return dateRef(t).Goize(), nil
	case C.CFArrayGetTypeID():
		return arrayRef(t).Goize()
	case C.CFDictionaryGetTypeID():
		return dictionaryRef(t).Goize()
	}
	return nil, &UnknownCFTypeError{typeId}
}

type stringRef C.CFStringRef

func (s stringRef) Goize() string {
	data := C.CFStringCreateExternalRepresentation(0, C.CFStringRef(s),
		C.kCFStringEncodingUTF8, 0)
	defer C.CFRelease(C.CFTypeRef(data))
	if data == 0 {
		panic("Unable to represent a CFString in UTF-8")
	}
	dataPtr := (*C.char)(unsafe.Pointer(C.CFDataGetBytePtr(data)))
	dataLen := C.int(C.CFDataGetLength(data))
	return C.GoStringN(dataPtr, dataLen)
}

type boolRef C.CFBooleanRef

func (b boolRef) Goize() bool {
	return C.CFBooleanGetValue(C.CFBooleanRef(b)) != 0
}

type numberRef C.CFNumberRef

func (n numberRef) GoizeFloat64() float64 {
	var v C.double
	C.CFNumberGetValue(C.CFNumberRef(n), C.kCFNumberDoubleType, unsafe.Pointer(&v))
	return float64(v)
}

func (n numberRef) GoizeInt64() int64 {
	var v C.SInt64
	C.CFNumberGetValue(C.CFNumberRef(n), C.kCFNumberSInt64Type, unsafe.Pointer(&v))
	return int64(v)
}

func (n numberRef) GoizeUint32() uint32 {
	var v C.SInt64
	C.CFNumberGetValue(C.CFNumberRef(n), C.kCFNumberSInt64Type, unsafe.Pointer(&v))
	return uint32(v)
}

func (n numberRef) Goize() interface{} {
	cfn := C.CFNumberRef(n)
	typ := C.CFNumberGetType(cfn)
	switch typ {
	case C.kCFNumberSInt8Type:
		var sint C.SInt8
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&sint))
		return int8(sint)
	case C.kCFNumberSInt16Type:
		var sint C.SInt16
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&sint))
		return int16(sint)
	case C.kCFNumberSInt32Type:
		var sint C.SInt32
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&sint))
		return int32(sint)
	case C.kCFNumberSInt64Type:
		return n.GoizeInt64()
	case C.kCFNumberFloat32Type:
		var float C.Float32
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&float))
		return float32(float)
	case C.kCFNumberFloat64Type:
		return n.GoizeFloat64()
	case C.kCFNumberCharType:
		var char C.char
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&char))
		return byte(char)
	case C.kCFNumberShortType:
		var short C.short
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&short))
		return int16(short)
	case C.kCFNumberIntType:
		var i C.int
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&i))
		return int32(i)
	case C.kCFNumberLongType:
		var long C.long
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&long))
		return int(long)
	case C.kCFNumberLongLongType:
		// this is the only type that may actually overflow us
		var longlong C.longlong
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&longlong))
		return int64(longlong)
	case C.kCFNumberFloatType:
		var float C.float
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&float))
		return float32(float)
	case C.kCFNumberDoubleType:
		var double C.double
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&double))
		return float64(double)
	case C.kCFNumberCFIndexType:
		// CFIndex is a long
		var index C.CFIndex
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&index))
		return int(index)
	case C.kCFNumberNSIntegerType:
		// We don't have a definition of NSInteger, but we know it's either an int or a long
		var nsInt C.long
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&nsInt))
		return int(nsInt)
	case C.kCFNumberCGFloatType:
		// CGFloat is a float or double
		var float C.CGFloat
		C.CFNumberGetValue(cfn, typ, unsafe.Pointer(&float))
		if unsafe.Sizeof(float) == 8 {
			return float64(float)
		} else {
			return float32(float)
		}
	}
	panic("plist: unknown CFNumber type")
}

type dataRef C.CFDataRef

func (d dataRef) Goize() []byte {
	bytes := C.CFDataGetBytePtr(C.CFDataRef(d))
	length := C.CFDataGetLength(C.CFDataRef(d))
	return C.GoBytes(unsafe.Pointer(bytes), C.int(length))
}

type dateRef C.CFDateRef

func (d dateRef) Goize() time.Time {
	nano := C.double(C.CFDateGetAbsoluteTime(C.CFDateRef(d)))
	nano += C.double(C.kCFAbsoluteTimeIntervalSince1970)
	// pull out milliseconds, to get a more predictable conversion
	ms := int64(float64(C.round(nano * 1000)))
	sec := ms / 1000
	nsec := (ms % 1000) * int64(time.Millisecond)
	return time.Unix(sec, nsec)
}

type arrayRef C.CFArrayRef

func (a arrayRef) Goize() ([]interface{}, error) {
	count := C.CFArrayGetCount(C.CFArrayRef(a))
	if count == 0 {
		return nil, nil
	}
	values := make([]C.CFTypeRef, int(count))
	out := make([]interface{}, int(count))
	C.CFArrayGetValues(C.CFArrayRef(a), C.CFRange{0, count}, (*unsafe.Pointer)(unsafe.Pointer(&values[0])))
	for i, value := range values {
		goValue, err := typeRef(value).Goize()
		if err != nil {
			return nil, err
		}
		out[i] = goValue
	}
	return out, nil
}

type dictionaryRef C.CFDictionaryRef

func (d dictionaryRef) Goize() (map[string]interface{}, error) {
	count := int(C.CFDictionaryGetCount(C.CFDictionaryRef(d)))
	if count == 0 {
		return map[string]interface{}{}, nil
	}
	stringTypeID := C.CFStringGetTypeID()
	keys := make([]C.CFTypeRef, count)
	values := make([]C.CFTypeRef, count)
	out := map[string]interface{}{}
	C.CFDictionaryGetKeysAndValues(C.CFDictionaryRef(d), (*unsafe.Pointer)(unsafe.Pointer(&keys[0])),
		(*unsafe.Pointer)(unsafe.Pointer(&values[0])))
	for i := 0; i < count; i++ {
		t := C.CFGetTypeID(keys[i])
		if t != stringTypeID {
			return nil, &UnsupportedKeyTypeError{int(t)}
		}
		key := stringRef(keys[i]).Goize()
		val, err := typeRef(values[i]).Goize()
		if err != nil {
			return nil, err
		}
		out[key] = val
	}
	return out, nil
}
