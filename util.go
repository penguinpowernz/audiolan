package audiolan

import (
	"bytes"
	"encoding/binary"
	"log"
	"reflect"
)

// Equal returns true if two audio packets are the same
func Equal(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Verify will verify a audio stream packet was encoded correctly
func Verify(buffer []float32, data []byte) bool {
	chkbuf := make([]float32, SampleRate*1)
	binary.Read(bytes.NewReader(data), binary.BigEndian, chkbuf)
	same := reflect.DeepEqual(buffer, chkbuf)
	eql := Equal(buffer, chkbuf)
	log.Println("are the same?", same)
	log.Println("are the eql?", eql)
	return same && eql
}
