package cf

// Taken from go-osx-plist (see LICENSE), heavily adapted

// #include <CoreFoundation/CoreFoundation.h>
import "C"
import (
	"reflect"
	"strconv"
)

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "plist: unsupported type: " + e.Type.String()
}

type UnknownCFTypeError struct {
	CFTypeID C.CFTypeID
}

func (e *UnknownCFTypeError) Error() string {
	cfStr := C.CFCopyTypeIDDescription(e.CFTypeID)
	str := stringRef(cfStr).Goize()
	Release(cfStr)
	return "plist: unknown CFTypeID " + strconv.Itoa(int(e.CFTypeID)) + " (" + str + ")"
}

type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "json: unsupported value: " + e.Str
}

type UnsupportedKeyTypeError struct {
	CFTypeID int
}

func (e *UnsupportedKeyTypeError) Error() string {
	return "plist: unexpected dictionary key CFTypeID " + strconv.Itoa(e.CFTypeID)
}
