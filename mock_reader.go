package audiolan

import "strconv"

type MockReader struct {
	i int
}

func (r *MockReader) Read(p []byte) (n int, err error) {
	r.i++
	buf := []byte(strconv.Itoa(r.i))
	n = copy(p, buf)
	return
}
