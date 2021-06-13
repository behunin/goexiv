package goexiv_test

import (
	"io/ioutil"
	"testing"

	"github.com/behunin/goexiv"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestOpenImage(t *testing.T) {
	// Open valid file
	img, err := goexiv.Open("testdata/pixel.jpg")

	if err != nil {
		t.Fatalf("Cannot open image: %s", err)
	}

	defer img.Close()

	if img == nil {
		t.Fatalf("img is nil after successful open")
	}

	// Open non existing file

	_, err = goexiv.Open("thisimagedoesnotexist")

	if err == nil {
		t.Fatalf("No error set after opening a non existing image")
	}

	exivErr, ok := err.(*goexiv.Error)

	if !ok {
		t.Fatalf("Returned error is not of type Error")
	}

	if exivErr.Code() != 9 {
		t.Fatalf("Unexpected error code (expected 9, got %d)", exivErr.Code())
	}
}

func Test_OpenBytes(t *testing.T) {
	bytes, err := ioutil.ReadFile("testdata/pixel.jpg")
	assert.NilError(t, err)

	img, err := goexiv.OpenBytes(bytes)
	assert.NilError(t, err)
	defer img.Close()
	assert.Check(t, nil != img)
}

func Test_OpenBytesFailures(t *testing.T) {
	tests := []struct {
		name        string
		bytes       []byte
		wantErr     string
		wantErrCode int
	}{
		{
			"no image",
			[]byte("no image"),
			"The memory contains data of an unknown image type",
			12,
		},
		{
			"empty byte slice",
			[]byte{},
			"input is empty",
			0,
		},
		{
			"nil byte slice",
			nil,
			"input is empty",
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := goexiv.OpenBytes(tt.bytes)
			if cmp.ErrorContains(err, tt.wantErr)().Success() {
				exivErr, ok := err.(*goexiv.Error)
				if !ok {
					assert.Equal(t, tt.wantErrCode, exivErr.Code(), "unexpected error code")
				}
			}
		})
	}
}

func TestIPTC(t *testing.T) {
	img, err := goexiv.Open("testdata/pixel.jpg")

	assert.NilError(t, err)
	defer img.Close()

	err = img.ReadMetadata()

	if err != nil {
		t.Fatalf("Cannot read image metadata: %s", err)
	}

	width := img.PixelWidth()
	height := img.PixelHeight()
	if width != 1 || height != 1 {
		t.Errorf("Cannot read image size (expected 1x1, got %dx%d)", width, height)
	}

	iptcData := img.GetIptcData()

	assert.Check(t, iptcData != nil)

	// Iterate over all IPTC data accessing Key() and String()
	keyValues := map[string]string{}
	i := iptcData.Iterator()
	for i.HasNext() {
		d := i.Next()
		keyValues[d.Key()] = d.String()
	}
	i.Close()
	iptcData.Close()
	assert.DeepEqual(t, keyValues, map[string]string{
		"Iptc.Application2.Caption":     "Hello, world! Привет, мир!",
		"Iptc.Application2.Copyright":   "this is the copy, right?",
		"Iptc.Application2.CountryName": "Lancre",
		"Iptc.Application2.DateCreated": "1848-10-13",
		"Iptc.Application2.TimeCreated": "12:49:32+01:00",
	})
}

func TestExif(t *testing.T) {
	img, err := goexiv.Open("testdata/pixel.jpg")

	assert.NilError(t, err)
	defer img.Close()

	err = img.ReadMetadata()

	if err != nil {
		t.Fatalf("Cannot read image metadata: %s", err)
	}

	width := img.PixelWidth()
	height := img.PixelHeight()
	if width != 1 || height != 1 {
		t.Errorf("Cannot read image size (expected 1x1, got %dx%d)", width, height)
	}

	data := img.GetExifData()

	if data == nil {
		t.Fatalf("Data Not Found!")
	}

	// Invalid key
	datum, err := data.FindKey("NotARealKey")

	if err == nil {
		t.Fatalf("FindKey returns a nil error for an invalid key")
	}

	if datum != nil {
		t.Fatalf("FindKey does not return nil for an invalid key")
	}

	// Valid, existing key

	datum, err = data.FindKey("Exif.Image.Make")

	if err != nil {
		t.Fatalf("FindKey returns an error for a valid, existing key: %s", err)
	}

	if datum == nil {
		t.Fatalf("FindKey returns nil for a valid, existing key")
	}

	if datum.String() != "FakeMake" {
		t.Fatalf("Unexpected value for EXIF datum Exif.Image.Make (expected 'FakeMake', got '%s')", datum.String())
	}

	// Valid, non existing key

	datum, err = data.FindKey("Exif.Photo.Flash")

	if err == nil {
		t.Fatalf("FindKey should return an error for a valid, non existing key: %s", err)
	}

	if datum != nil {
		t.Fatalf("FindKey returns a non null datum for a valid, non existing key")
	}

	// Iterate over all Exif data accessing Key() and String()
	keyValues := map[string]string{}
	i := data.Iterator()
	for i.HasNext() {
		d := i.Next()
		keyValues[d.Key()] = d.String()
	}
	i.Close()
	data.Close()
	assert.DeepEqual(t, keyValues, map[string]string{
		"Exif.Image.ExifTag":                 "134",
		"Exif.Image.Make":                    "FakeMake",
		"Exif.Image.Model":                   "FakeModel",
		"Exif.Image.ResolutionUnit":          "2",
		"Exif.Image.XResolution":             "72/1",
		"Exif.Image.YCbCrPositioning":        "1",
		"Exif.Image.YResolution":             "72/1",
		"Exif.Photo.ColorSpace":              "65535",
		"Exif.Photo.ComponentsConfiguration": "1 2 3 0",
		"Exif.Photo.DateTimeDigitized":       "2013:12:08 21:06:10",
		"Exif.Photo.ExifVersion":             "48 50 51 48",
		"Exif.Photo.FlashpixVersion":         "48 49 48 48",
	})
}

func TestXMP(t *testing.T) {
	img, err := goexiv.Open("testdata/OZZY.jpg")

	assert.NilError(t, err)
	defer img.Close()

	err = img.ReadMetadata()

	if err != nil {
		t.Fatalf("Cannot read image metadata: %s", err)
	}

	data := img.GetXmpData()

	if data == nil {
		t.Fatalf("Data Not Found!")
	}

	// Valid, existing key

	datum, err := data.FindKey("Xmp.xmpRights.WebStatement")

	if err != nil {
		t.Fatalf("FindKey returns an error for an valid key")
	}

	if datum == nil {
		t.Fatalf("FindKey does return nil for an valid key")
	}

	// Invalid key
	datum, err = data.FindKey("NotARealKey")

	if err == nil {
		t.Fatalf("FindKey returns a nil error for an invalid key")
	}

	if datum != nil {
		t.Fatalf("FindKey does not return nil for an invalid key")
	}

	// Iterate over all Exif data accessing Key() and String()
	keyValues := map[string]string{}
	i := data.Iterator()
	for i.HasNext() {
		d := i.Next()
		keyValues[d.Key()] = d.String()
	}
	i.Close()
	data.Close()
	assert.DeepEqual(t, keyValues, map[string]string{
		"Xmp.xmpRights.WebStatement":     "iRockimages.com",
		"Xmp.xmpRights.UsageTerms":       "",
		"Xmp.photoshop.LegacyIPTCDigest": "67B99F1727D6D51B794526CA4D330C86",
		"Xmp.tiff.ImageWidth":            "433",
		"Xmp.tiff.ImageLength":           "650",
		"Xmp.tiff.XResolution":           "240/1",
		"Xmp.tiff.YResolution":           "240/1",
		"Xmp.tiff.ResolutionUnit":        "2",
		"Xmp.dc.rights":                  "iRockimages.com",
	})
}

func TestNoICC(t *testing.T) {
	img, err := goexiv.Open("testdata/stripped_pixel.jpg")
	assert.NilError(t, err)
	err = img.ReadMetadata()
	assert.NilError(t, err)

	assert.Assert(t, cmp.Nil(img.ICCProfile()))
}

type MetadataTestCase struct {
	Format                 string // exif or iptc
	Key                    string
	Value                  string
	ImageFilename          string
	ExpectedErrorSubstring string
}

func TestSetMetadataString(t *testing.T) {
	cases := []MetadataTestCase{
		// valid exif key, jpeg
		{
			Format:                 "exif",
			Key:                    "Exif.Photo.UserComment",
			Value:                  "Hello, world! Привет, мир!",
			ImageFilename:          "testdata/stripped_pixel.jpg",
			ExpectedErrorSubstring: "", // no error
		},
		// valid exif key, webp
		{
			Format:                 "exif",
			Key:                    "Exif.Photo.UserComment",
			Value:                  "Hello, world! Привет, мир!",
			ImageFilename:          "testdata/pixel.webp",
			ExpectedErrorSubstring: "",
		},
		// valid iptc key, jpeg.
		// webp iptc is not supported (see libexiv2/src/webpimage.cpp WebPImage::setIptcData))
		{
			Format:                 "iptc",
			Key:                    "Iptc.Application2.Caption",
			Value:                  "Hello, world! Привет, мир!",
			ImageFilename:          "testdata/stripped_pixel.jpg",
			ExpectedErrorSubstring: "",
		},
		// invalid exif key, jpeg
		{
			Format:                 "exif",
			Key:                    "Exif.Invalid.Key",
			Value:                  "this value should not be written",
			ImageFilename:          "testdata/stripped_pixel.jpg",
			ExpectedErrorSubstring: "Invalid key",
		},
		// invalid exif key, webp
		{
			Format:                 "exif",
			Key:                    "Exif.Invalid.Key",
			Value:                  "this value should not be written",
			ImageFilename:          "testdata/pixel.webp",
			ExpectedErrorSubstring: "Invalid key",
		},
		// invalid iptc key, jpeg
		{
			Format:                 "iptc",
			Key:                    "Iptc.Invalid.Key",
			Value:                  "this value should not be written",
			ImageFilename:          "testdata/stripped_pixel.jpg",
			ExpectedErrorSubstring: "Invalid record name",
		},
	}

	var data goexiv.MetadataProvider

	for i, testcase := range cases {
		img, err := goexiv.Open(testcase.ImageFilename)
		assert.NilError(t, err, "case #%d Error while opening image file", i)

		err = img.SetMetadataString(testcase.Format, testcase.Key, testcase.Value)
		if testcase.ExpectedErrorSubstring != "" {
			assert.ErrorContains(
				t,
				err,
				testcase.ExpectedErrorSubstring,
				"case #%d Error text must contain a given substring",
				i,
			)
		} else {
			assert.NilError(t, err, "case #%d Cannot write image metadata", i)
		}

		err = img.ReadMetadata()
		assert.NilError(t, err, "case #%d Cannot read image metadata", i)

		if testcase.Format == "iptc" {
			data = img.GetIptcData()
		} else {
			data = img.GetExifData()
		}

		receivedValue, err := data.GetString(testcase.Key)
		data.Close()
		if err != nil {
			assert.ErrorContains(
				t,
				err,
				testcase.ExpectedErrorSubstring,
				"case #%d must contain %s: %s given",
				i,
				testcase.ExpectedErrorSubstring,
				receivedValue,
			)
		}
		img.Close()
	}
}

func Test_GetBytes(t *testing.T) {
	bytes, err := ioutil.ReadFile("testdata/stripped_pixel.jpg")
	assert.NilError(t, err)

	img, err := goexiv.OpenBytes(bytes)
	assert.NilError(t, err)

	assert.Equal(
		t,
		len(bytes),
		len(img.GetBytes()),
		"Image size on disk and in memory must be equal",
	)

	bytesBeforeTag := img.GetBytes()
	assert.NilError(t, img.SetExifString("Exif.Photo.UserComment", "123"))
	bytesAfterTag := img.GetBytes()
	assert.Check(t, len(bytesAfterTag) > len(bytesBeforeTag), "Image size must increase after adding an EXIF tag")
	assert.DeepEqual(t, &bytesBeforeTag[0], &bytesAfterTag[0])

	assert.NilError(t, img.SetExifString("Exif.Photo.UserComment", "123"))
	bytesAfterTag2 := img.GetBytes()
	assert.Equal(
		t,
		len(bytesAfterTag),
		len(bytesAfterTag2),
		"Image size must not change after the same tag has been set",
	)
}
