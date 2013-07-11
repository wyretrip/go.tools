// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements tests for various issues.

package types

import (
	"go/ast"
	"go/parser"
	"strings"
	"testing"

	"code.google.com/p/go.tools/go/exact"
)

func TestIssue5770(t *testing.T) {
	src := `package p; type S struct{T}`
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = Check(f.Name.Name, fset, f) // do not crash
	want := "undeclared name: T"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("got: %v; want: %s", err, want)
	}
}

func TestIssue5849(t *testing.T) {
	src := `
package p
var (
	s uint
	_ = uint8(8)
	_ = uint16(16) << s
	_ = uint32(32 << s)
	_ = uint64(64 << s + s)
	_ = (interface{})("foo")
	_ = (interface{})(nil)
)`
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Error(err)
		return
	}

	ctxt := Context{
		Expr: func(x ast.Expr, typ Type, val exact.Value) {
			var want Type
			switch x := x.(type) {
			case *ast.BasicLit:
				switch x.Value {
				case `8`:
					want = Typ[Uint8]
				case `16`:
					want = Typ[Uint16]
				case `32`:
					want = Typ[Uint32]
				case `64`:
					want = Typ[Uint] // because of "+ s", s is of type uint
				case `"foo"`:
					want = Typ[String]
				}
			case *ast.Ident:
				if x.Name == "nil" {
					want = Typ[UntypedNil]
				}
			}
			if want != nil && !IsIdentical(typ, want) {
				t.Errorf("got %s; want %s", typ, want)
			}
		},
	}

	_, err = ctxt.Check(f.Name.Name, fset, f)
	if err != nil {
		t.Error(err)
	}
}