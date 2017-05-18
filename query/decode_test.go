package query

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
)

type User struct {
	Name string `url:"name"`
	Addr Addr   `url:"addr"`
}

type Addr struct {
	ZipCode int    `url:"zipcode"`
	City    string `url:"city"`
}

type A struct {
	Bool             *bool     `url:"bool"`
	Int              int       `url:"int"`
	Float32          float32   `url:"float32"`
	Float64          float64   `url:"float64"`
	Str              string    `url:"str"`
	ArrStr           [2]string `url:"arrStr"`
	SliceStr         []string  `url:"sliceStr"`
	SliceStrBrackets []string  `url:"sliceStrBrackets,brackets"`
	SliceSepStr      []string  `url:"sliceSepStr,comma"`
	User             User      `url:"user"`
}

type B struct {
	X interface{} `url:"x"`
}

func TestDecode(t *testing.T) {
	cases := []struct {
		values    url.Values
		ptr       interface{}
		want      interface{}
		wantError bool
	}{
		{
			ptr:       nil,
			wantError: true,
		},
		{
			values: url.Values{
				"bool":                {"false"},
				"int":                 {"42"},
				"float32":             {"32"},
				"float64":             {"314E-2"},
				"str":                 {"str"},
				"arrStr":              {"A", "B"},
				"sliceStr":            {"X", "Y", "Z"},
				"sliceStrBrackets[]":  {"Xi", "Yi", "Zi"},
				"sliceSepStr":         {"Woo,Xoo,Yoo,Zoo"},
				"user[name]":          {"acme"},
				"user[addr][zipcode]": {"1234"},
				"user[addr][city]":    {"SFO"},
			},
			ptr: &A{},
			want: &A{
				Bool:             new(bool),
				Int:              42,
				Float32:          float32(32),
				Float64:          float64(3.14),
				Str:              "str",
				ArrStr:           [2]string{"A", "B"},
				SliceStr:         []string{"X", "Y", "Z"},
				SliceStrBrackets: []string{"Xi", "Yi", "Zi"},
				SliceSepStr:      []string{"Woo", "Xoo", "Yoo", "Zoo"},
				User: User{
					Name: "acme",
					Addr: Addr{
						ZipCode: 1234,
						City:    "SFO",
					},
				},
			},
		},
		{
			values: url.Values{
				"x[name]":          {"Xacme"},
				"x[addr][zipcode]": {"12345"},
				"x[addr][city]":    {"XSFO"},
			},
			ptr: &B{X: new(User)},
			want: &B{
				X: &User{
					Name: "Xacme",
					Addr: Addr{
						ZipCode: 12345,
						City:    "XSFO",
					},
				},
			},
		},
	}

	for i, tc := range cases {
		got := tc.ptr
		err := Decode(tc.values, got)
		if err != nil {
			if !tc.wantError {
				t.Errorf("#%d: Decode(%#v, ptr) returns error %s", i, tc.values, err)
			}
			continue
		}
		if tc.wantError {
			t.Errorf("#%d: Decode requires error", i)
			continue
		}
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("#%d: Decode mismatch, have %#v, want %#v", i, got, tc.want)
		}
		if b, ok := got.(*B); ok {
			fmt.Printf("got: %#v\n", b.X)
		}

		if b, ok := tc.want.(*B); ok {
			fmt.Printf("want: %#v\n", b.X)
		}
	}
}
