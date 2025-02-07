package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

type ExifData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2ExifData
}

type ExifDatum struct {
	data  *ExifData
	datum *C.Exiv2ExifDatum
}

// ExifDatumIterator wraps the respective C++ structure.
type ExifDatumIterator struct {
	data *ExifData
	iter *C.Exiv2ExifDatumIterator
}

func makeExifData(img *Image, cdata *C.Exiv2ExifData) *ExifData {
	if img == nil || cdata == nil {
		return nil
	}

	return &ExifData{
		img,
		cdata,
	}
}

func makeExifDatum(data *ExifData, cdatum *C.Exiv2ExifDatum) *ExifDatum {
	if data == nil || cdatum == nil {
		return nil
	}

	return &ExifDatum{
		data,
		cdatum,
	}
}

func (i *Image) GetExifData() *ExifData {
	return makeExifData(i, C.exiv2_image_get_exif_data(i.img))
}

func (i *Image) SetExifString(key, value string) error {
	return i.SetMetadataString("exif", key, value)
}

// Close free's the Exif data.
func (d *ExifData) Close() {
	C.exiv2_exif_data_free(d.data)
}

func (d *ExifData) GetString(key string) (string, error) {
	datum, err := d.FindKey(key)
	if err != nil {
		return "", err
	}

	if datum == nil {
		return "", ErrMetadataKeyNotFound
	}

	return datum.String(), nil
}

func (d *ExifData) FindKey(key string) (*ExifDatum, error) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var cerr *C.Exiv2Error

	cdatum := C.exiv2_exif_data_find_key(d.data, ckey, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	return makeExifDatum(d, cdatum), nil
}

// Key returns the Exif key of the datum.
func (d *ExifDatum) Key() string {
	return C.GoString(C.exiv2_exif_datum_key(d.datum))
}

func (d *ExifDatum) String() string {
	cstr := C.exiv2_exif_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}

// Iterator returns a new ExifDatumIterator to iterate over all Exif data.
func (d *ExifData) Iterator() *ExifDatumIterator {
	return makeExifDatumIterator(d, C.exiv2_exif_data_iterator(d.data))
}

// HasNext returns true as long as the iterator has another datum to deliver.
func (i *ExifDatumIterator) HasNext() bool {
	return C.exiv2_exif_data_iterator_has_next(i.iter) != 0
}

// Next returns the next ExifDatum of the iterator or nil if iterator has reached the end.
func (i *ExifDatumIterator) Next() *ExifDatum {
	return makeExifDatum(i.data, C.exiv2_exif_datum_iterator_next(i.iter))
}

// Close free's the Exif datum iterator.
func (i *ExifDatumIterator) Close() {
	C.exiv2_exif_datum_iterator_free(i.iter)
}

func makeExifDatumIterator(data *ExifData, cIter *C.Exiv2ExifDatumIterator) *ExifDatumIterator {
	return &ExifDatumIterator{data, cIter}
}
