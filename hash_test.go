package dht

import "testing"

func Test_newHash(t *testing.T) {
	s1 := "574a7773ca432765f6da014e7a42d1ff1572a458"
	testNewHash(t, s1)
	s2 := "574a7773ca432765f6da014e7a42d1ff1572a450"
	testNewHash(t, s2)
}

func testNewHash(t *testing.T, s string) {
	h, err := newHash([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	if h.String() != s {
		t.Fatal(h)
	}
}

func Test_hash_compare(t *testing.T) {
	s1 := "574a7773ca432765f6da014e7a42d1ff1572a458"
	s2 := "574a7773ca432765f6da014e7a42d1ff1572a450"
	testHashCompare(t, s1, s2, 1)

	s3 := "574a7773ca432765f6da014e7a42d1ff1572a450"
	s4 := "574a7773ca432765f6da014e7a42d1ff1572a450"
	testHashCompare(t, s3, s4, 0)

	s5 := "174a7773ca432765f6da014e7a42d1ff1572a450"
	s6 := "574a7773ca432765f6da014e7a42d1ff1572a450"
	testHashCompare(t, s5, s6, -1)
}

func testHashCompare(t *testing.T, s1, s2 string, n int) {
	ss := []string{"<", "=", ">"}
	h1, err := newHash([]byte(s1))
	if err != nil {
		t.Fatal(err)
	}
	h2, err := newHash([]byte(s2))
	if err != nil {
		t.Fatal(err)
	}
	n1 := h1.compare(h2)
	if n1 != n {
		t.Fatalf("%s %s %s", h1, h2, ss[n+1])
	}
}

func Test_hash_middle(t *testing.T) {
	s1 := "574a7773ca432765f6da014e7a42d1ff1572a458"
	s2 := "574a7773ca432765f6da014e7a42d1ff1572a450"
	s3 := "574a7773ca432765f6da014e7a42d1ff1572a454"
	testHashMiddle(t, s1, s2, s3)
	testHashMiddle(t, s2, s1, s3)
}

func testHashMiddle(t *testing.T, s1, s2, s3 string) {
	h1, err := newHash([]byte(s1))
	if err != nil {
		t.Fatal(err)
	}
	h2, err := newHash([]byte(s2))
	if err != nil {
		t.Fatal(err)
	}
	h3 := h1.middle(h2)
	if h3.String() != s3 {
		t.Fatalf("%s != %s", h3.String(), s3)
	}
}
