package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)

// XmpData contains all Xmp Data of an image.
type XmpData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2XmpData
}

// XmpDatum stores the info of one xmp datum.
type XmpDatum struct {
	data  *XmpData
	datum *C.Exiv2XmpDatum
}

// XmpDatumIterator wraps the respective C++ structure.
type XmpDatumIterator struct {
	data *XmpData
	iter *C.Exiv2XmpDatumIterator
}

func makeXmpData(img *Image, cdata *C.Exiv2XmpData) *XmpData {
	if img == nil || cdata == nil {
		return nil
	}

	return &XmpData{
		img,
		cdata,
	}
}

func makeXmpDatum(data *XmpData, cdatum *C.Exiv2XmpDatum) *XmpDatum {
	if data == nil || cdatum == nil {
		return nil
	}

	return &XmpDatum{
		data,
		cdatum,
	}
}

// GetXmpData returns the XmpData of an Image.
func (i *Image) GetXmpData() *XmpData {
	return makeXmpData(i, C.exiv2_image_get_xmp_data(i.img))
}

// Close free's the Xmp data.
func (d *XmpData) Close() {
	C.exiv2_xmp_data_free(d.data)
}

// FindKey tries to find the specified key and returns its data.
// It returns an error if the key is invalid. If the key is not found, a
// nil pointer will be returned
func (d *XmpData) FindKey(key string) (*XmpDatum, error) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var cerr *C.Exiv2Error

	cdatum := C.exiv2_xmp_data_find_key(d.data, ckey, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	runtime.KeepAlive(d)
	return makeXmpDatum(d, cdatum), nil
}

// Key returns the Xmp key of the datum.
func (d *XmpDatum) Key() string {
	cstr := C.exiv2_xmp_datum_key(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}

func (d *XmpDatum) String() string {
	cstr := C.exiv2_xmp_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}

func (d *XmpDatum) Type() string {
	cstr := C.exiv2_xmp_datum_type(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}

// Iterator returns a new ExifDatumIterator to iterate over all Exif data.
func (d *XmpData) Iterator() *XmpDatumIterator {
	return makeXmpDatumIterator(d, C.exiv2_xmp_data_iterator(d.data))
}

// HasNext returns true as long as the iterator has another datum to deliver.
func (i *XmpDatumIterator) HasNext() bool {
	return C.exiv2_xmp_data_iterator_has_next(i.iter) != 0
}

// Next returns the next ExifDatum of the iterator or nil if iterator has reached the end.
func (i *XmpDatumIterator) Next() *XmpDatum {
	return makeXmpDatum(i.data, C.exiv2_xmp_datum_iterator_next(i.iter))
}

func (i *XmpDatumIterator) Close() {
	C.exiv2_xmp_datum_iterator_free(i.iter)
}

func makeXmpDatumIterator(data *XmpData, cIter *C.Exiv2XmpDatumIterator) *XmpDatumIterator {
	return &XmpDatumIterator{data, cIter}
}
