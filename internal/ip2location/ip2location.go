package ip2location

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"

	"github.com/pomerium/datasource/internal/jsonutil"
	"github.com/pomerium/datasource/internal/netutil"
)

func fileToJSON(dst *jsonutil.JSONArrayStream, fileName string) (err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	mt, err := mimetype.DetectReader(f)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	if mt.Is("application/zip") {
		zr, err := zip.NewReader(f, fi.Size())
		if err != nil {
			return err
		}
		return zipToJSON(dst, zr)
	}

	return csvToJSON(dst, f)
}

func zipToJSON(dst *jsonutil.JSONArrayStream, zr *zip.Reader) (err error) {
	for _, zf := range zr.File {
		if filepath.Ext(strings.ToLower(zf.Name)) == ".csv" {
			rc, err := zf.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			return csvToJSON(dst, rc)
		}
	}
	return fmt.Errorf("no csv file found in zip file")
}

func csvToJSON(dst *jsonutil.JSONArrayStream, r io.Reader) error {
	cr := csv.NewReader(r)
	for {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		err = csvRowToJSON(dst, row)
		if err != nil {
			return err
		}
	}
	return dst.Close()
}

// parseCSVRow parses an individual row of the  CSV file.
func csvRowToJSON(dst *jsonutil.JSONArrayStream, row []string) (err error) {
	if len(row) < 2 {
		return nil
	}

	start, err := netutil.ParseIPNumber(row[0])
	if err != nil {
		return err
	}

	end, err := netutil.ParseIPNumber(row[1])
	if err != nil {
		return err
	}

	cidrs := netutil.AddrRangeToPrefixes(start, end)
	for _, cidr := range cidrs {
		var record Record
		record.Index.CIDR = cidr.String()
		record.ID = cidr.String()
		if len(row) >= 3 {
			record.Country = row[2]
		}
		if len(row) >= 5 {
			record.State = row[4]
		}
		if len(row) >= 6 {
			record.City = row[5]
		}
		if len(row) >= 8 {
			record.Zip = row[8]
		}
		if len(row) >= 9 {
			record.Timezone = row[9]
		}
		// ignore rows not associated with a country
		if record.Country == "" || record.Country == "-" {
			continue
		}

		err = dst.Encode(record)
		if err != nil {
			return err
		}
	}

	return nil
}
