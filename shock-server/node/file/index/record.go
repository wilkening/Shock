package index

import (
	"encoding/binary"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/multi"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"io"
	"math/rand"
	"os"
)

type record struct {
	f     *os.File
	r     seq.Reader
	t     string
	snf   string
	snp   string
	Index *Idx
}

func NewRecordIndexer(f *os.File, nType string, snFormat string, snIndexPath string) Indexer {
	return &record{
		f:     f,
		r:     multi.NewReader(f),
		t:     nType,
		snf:   snFormat,
		snp:   snIndexPath,
		Index: New(),
	}
}

func (i *record) Create(file string) (count int64, format string, err error) {
	tmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.PATH_DATA, rand.Int(), rand.Int())

	f, err := os.Create(tmpFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	format = "array"
	eof := false
	curr := int64(0)
	count = 0
	buffer_pos := 0 // used to track the location in our byte array

	// Writing index file in 16MB chunks
	var b [16777216]byte
	for {
		n, er := i.r.GetReadOffset()
		if er != nil {
			if er != io.EOF {
				err = er
				return
			}
			eof = true
		}

		if eof && (n == 0) {
			break
		}

		x := (buffer_pos * 16)
		if x == 16777216 {
			f.Write(b[:])
			buffer_pos = 0
			x = 0
		}
		y := x + 8
		z := x + 16

		binary.LittleEndian.PutUint64(b[x:y], uint64(curr))
		binary.LittleEndian.PutUint64(b[y:z], uint64(n))
		curr += int64(n)
		count += 1
		buffer_pos += 1

		if eof {
			break
		}
	}
	if buffer_pos != 0 {
		f.Write(b[:buffer_pos*16])
	}
	err = os.Rename(tmpFilePath, file)

	return
}

func (i *record) Close() (err error) {
	i.f.Close()
	return
}
