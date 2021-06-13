#include "helper.h"

#include <exiv2/image.hpp>
#include <exiv2/error.hpp>

#include <stdio.h>

#define DEFINE_STRUCT(name,wrapped_type,member_name) \
struct _##name { \
	_##name(wrapped_type member_name) \
		: member_name(member_name) {} \
	wrapped_type member_name; \
};

DEFINE_STRUCT(Exiv2ImageFactory, Exiv2::ImageFactory*, factory);
DEFINE_STRUCT(Exiv2Image, Exiv2::Image::AutoPtr, image);

DEFINE_STRUCT(Exiv2XmpData, const Exiv2::XmpData&, data);
DEFINE_STRUCT(Exiv2XmpDatum, const Exiv2::Xmpdatum&, datum);
struct _Exiv2XmpDatumIterator {
	_Exiv2XmpDatumIterator(Exiv2::XmpMetadata::const_iterator i, Exiv2::XmpMetadata::const_iterator e) : it(i), end(e) {}
	Exiv2::XmpMetadata::const_iterator it;
	Exiv2::XmpMetadata::const_iterator end;

	bool has_next() const;
	Exiv2XmpDatum *next();
};

DEFINE_STRUCT(Exiv2ExifData, const Exiv2::ExifData&, data);
DEFINE_STRUCT(Exiv2ExifDatum, const Exiv2::Exifdatum&, datum);
struct _Exiv2ExifDatumIterator {
	_Exiv2ExifDatumIterator(Exiv2::ExifMetadata::const_iterator i, Exiv2::ExifMetadata::const_iterator e) : it(i), end(e) {}
	Exiv2::ExifMetadata::const_iterator it;
	Exiv2::ExifMetadata::const_iterator end;

	bool has_next() const;
	Exiv2ExifDatum* next();
};

DEFINE_STRUCT(Exiv2IptcData, const Exiv2::IptcData&, data);
DEFINE_STRUCT(Exiv2IptcDatum, const Exiv2::Iptcdatum&, datum);
struct _Exiv2IptcDatumIterator {
	_Exiv2IptcDatumIterator(Exiv2::IptcMetadata::const_iterator i, Exiv2::IptcMetadata::const_iterator e) : it(i), end(e) {}
	Exiv2::IptcMetadata::const_iterator it;
	Exiv2::IptcMetadata::const_iterator end;

	bool has_next() const;
	Exiv2IptcDatum* next();
};

void exiv2_iptc_datum_iterator_free(Exiv2IptcDatumIterator *x) { delete x; };
void exiv2_exif_datum_iterator_free(Exiv2ExifDatumIterator *x) { delete x; };
void exiv2_xmp_datum_iterator_free(Exiv2XmpDatumIterator *x) { delete x; };

struct _Exiv2Error {
	_Exiv2Error(const Exiv2::Error &error);

	int code;
	char *what;
};

_Exiv2Error::_Exiv2Error(const Exiv2::Error &error)
	: code(error.code())
	, what(strdup(error.what()))
{
}

Exiv2Image*
exiv2_image_factory_open(const char *path, Exiv2Error **error)
{
	Exiv2Image *p = 0;

	try {
		p = new Exiv2Image(Exiv2::ImageFactory::open(path));
		return p;
	} catch (Exiv2::Error &e) {
		delete p;

		if (error) {
			*error = new Exiv2Error(e);
		}
	}

	return 0;
}

Exiv2Image*
exiv2_image_factory_open_bytes(const unsigned char *bytes, long size, Exiv2Error **error)
{
	Exiv2Image *p = 0;

	try {
		p = new Exiv2Image(Exiv2::ImageFactory::open(bytes, size));
		return p;
	} catch (Exiv2::Error &e) {
		delete p;

		if (error) {
			*error = new Exiv2Error(e);
		}
	}

	return 0;
}

void
exiv2_image_read_metadata(Exiv2Image *img, Exiv2Error **error)
{
	try {
		img->image->readMetadata();
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}
	}
}

void
exiv2_image_set_exif_string(Exiv2Image *img, char *key, char *value, Exiv2Error **error)
{
	Exiv2::ExifData exifData = img->image->exifData();

	try {
	    Exiv2::Exifdatum& tag = exifData[key];
		Exiv2::Value::AutoPtr valueObject = Exiv2::Value::create(Exiv2::asciiString);
		valueObject->read(value);
		tag.setValue(valueObject.get());

		img->image->setExifData(exifData);
		img->image->writeMetadata();
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}
	}
}

void
exiv2_image_set_iptc_string(Exiv2Image *img, char *key, char *value, Exiv2Error **error)
{
	Exiv2::IptcData iptcData = img->image->iptcData();

	try {
		Exiv2::StringValue valueObject;
		valueObject.read(value);
		iptcData[key] = valueObject;

		img->image->setIptcData(iptcData);
		img->image->writeMetadata();
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}
	}
}

long
exiv_image_get_size(Exiv2Image *img)
{
    return (long)img->image->io().size();
}

unsigned char*
exiv_image_get_bytes_ptr(Exiv2Image *img)
{
    return img->image->io().mmap();
}


void exiv2_image_free(Exiv2Image *x) { delete x; };

int exiv2_image_get_pixel_width(Exiv2Image *img) {
	return img->image->pixelWidth();
}

int exiv2_image_get_pixel_height(Exiv2Image *img) {
	return img->image->pixelHeight();
}

const unsigned char* exiv2_image_icc_profile(Exiv2Image *img)
{
	if (img->image->iccProfileDefined()) {
		return img->image->iccProfile()->pData_;
	}
	return NULL;
}

long exiv2_image_icc_profile_size(Exiv2Image *img)
{
	if (img->image->iccProfileDefined()) {
		return img->image->iccProfile()->size_;
	}
	return 0;
}

// XMP
Exiv2XmpData*
exiv2_image_get_xmp_data(const Exiv2Image *img)
{
	return new Exiv2XmpData(img->image->xmpData());
}

Exiv2XmpDatum*
exiv2_xmp_data_find_key(const Exiv2XmpData *data, const char *key, Exiv2Error **error)
{
	try {
		const Exiv2::XmpData::const_iterator it = data->data.findKey(Exiv2::XmpKey(key));
		if (it == data->data.end()) {
			return 0;
		}

		return new Exiv2XmpDatum(*it);
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}

		return 0;
	}
}

void exiv2_xmp_data_free(Exiv2XmpData *x) { delete x; };

char*
exiv2_xmp_datum_to_string(const Exiv2XmpDatum *datum)
{
    Exiv2::TypeId typeId = datum->datum.typeId();

    std::string strval;

    if (typeId == Exiv2::xmpBag) {
        strval = datum->datum.toString();
    } else {
        strval = datum->datum.toString(0);
    }

	return strdup(strval.c_str());
}

void exiv2_xmp_datum_free(Exiv2XmpDatum *x) { delete x; };

Exiv2XmpDatumIterator *exiv2_xmp_data_iterator(const Exiv2XmpData *data)
{
	return new Exiv2XmpDatumIterator(data->data.begin(), data->data.end());
}

bool Exiv2XmpDatumIterator::has_next() const { return it != end; }

int exiv2_xmp_data_iterator_has_next(const Exiv2XmpDatumIterator *iter)
{
	return iter->has_next() ? 1 : 0;
}

Exiv2XmpDatum *Exiv2XmpDatumIterator::next()
{
	if (it == end)
	{
		return 0;
	}
	return new Exiv2XmpDatum(*it++);
}

Exiv2XmpDatum *exiv2_xmp_datum_iterator_next(Exiv2XmpDatumIterator *iter)
{
	return iter->next();
}

const char *exiv2_xmp_datum_key(const Exiv2XmpDatum *datum)
{
	const std::string strkey = datum->datum.key();
	return strdup(strkey.c_str());
}

// IPTC

Exiv2IptcData*
exiv2_image_get_iptc_data(const Exiv2Image *img)
{
	return new Exiv2IptcData(img->image->iptcData());
}

Exiv2IptcDatum*
exiv2_iptc_data_find_key(const Exiv2IptcData *data, const char *key, Exiv2Error **error)
{
	try {
		const Exiv2::IptcData::const_iterator it = data->data.findKey(Exiv2::IptcKey(key));
		if (it == data->data.end()) {
			return 0;
		}

		return new Exiv2IptcDatum(*it);
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}

		return 0;
	}
}

Exiv2IptcDatumIterator* exiv2_iptc_data_iterator(const Exiv2IptcData *data)
{
	return new Exiv2IptcDatumIterator(data->data.begin(), data->data.end());
}

bool Exiv2IptcDatumIterator::has_next() const
{
	return it != end;
}

int exiv2_iptc_data_iterator_has_next(const Exiv2IptcDatumIterator *iter)
{
	return iter->has_next() ? 1 : 0;
}

Exiv2IptcDatum* Exiv2IptcDatumIterator::next()
{
	if (it == end) {
		return 0;
	}
	return new Exiv2IptcDatum(*it++);
}

Exiv2IptcDatum* exiv2_iptc_datum_iterator_next(Exiv2IptcDatumIterator *iter)
{
	return iter->next();
}

void exiv2_iptc_data_free(Exiv2IptcData *x) { delete x; };

const char* exiv2_iptc_datum_key(const Exiv2IptcDatum *datum)
{
	const std::string strkey = datum->datum.key();
	return strdup(strkey.c_str());
}

const char* exiv2_iptc_datum_to_string(const Exiv2IptcDatum *datum)
{
	const std::string strval = datum->datum.toString();
	return strdup(strval.c_str());
}

void exiv2_iptc_datum_free(Exiv2IptcDatum *x) { delete x; };

// EXIF

Exiv2ExifData*
exiv2_image_get_exif_data(const Exiv2Image *img)
{
	return new Exiv2ExifData(img->image->exifData());
}

Exiv2ExifDatum*
exiv2_exif_data_find_key(const Exiv2ExifData *data, const char *key, Exiv2Error **error)
{
	try {
		const Exiv2::ExifData::const_iterator it = data->data.findKey(Exiv2::ExifKey(key));
		if (it == data->data.end()) {
			return 0;
		}

		return new Exiv2ExifDatum(*it);
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}

		return 0;
	}
}

Exiv2ExifDatumIterator* exiv2_exif_data_iterator(const Exiv2ExifData *data)
{
	return new Exiv2ExifDatumIterator(data->data.begin(), data->data.end());
}

bool Exiv2ExifDatumIterator::has_next() const
{
	return it != end;
}

int exiv2_exif_data_iterator_has_next(const Exiv2ExifDatumIterator *iter)
{
	return iter->has_next() ? 1 : 0;
}

Exiv2ExifDatum* Exiv2ExifDatumIterator::next()
{
	if (it == end) {
		return 0;
	}
	return new Exiv2ExifDatum(*it++);
}

Exiv2ExifDatum* exiv2_exif_datum_iterator_next(Exiv2ExifDatumIterator *iter)
{
	return iter->next();
}

void exiv2_exif_data_free(Exiv2ExifData *x) { delete x; };

const char* exiv2_exif_datum_key(const Exiv2ExifDatum *datum)
{
	const std::string strkey = datum->datum.key();
	return strdup(strkey.c_str());
}

const char* exiv2_exif_datum_to_string(const Exiv2ExifDatum *datum)
{
	const std::string strval = datum->datum.toString();
	return strdup(strval.c_str());
}

void exiv2_exif_datum_free(Exiv2ExifDatum *x) { delete x; };

// ERRORS

int
exiv2_error_code(const Exiv2Error *error)
{
	return error->code;
}

const char*
exiv2_error_what(const Exiv2Error *error)
{
	return error->what;
}

void
exiv2_error_free(Exiv2Error *e)
{
	if (e == 0) {
		return;
	}

	free(e->what);
	delete e;
}
