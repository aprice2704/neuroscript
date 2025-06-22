// FDM Version: 0.2.0
// File version: 3
// Purpose: Print “type → interface” matrix (value *and* pointer) using go/packages
// filename: cmd/impliface/main.go
// nlines: ~160
// risk_rating: LOW

package main

import (
	"flag"
	"fmt"
	"go/types"
	"log"
	"os"
	"sort"

	"golang.org/x/tools/go/packages"
)

var allPkgs = flag.Bool("all", false, "load ./... instead of explicit patterns")

func main() {
	flag.Parse()
	patterns := flag.Args()
	if *allPkgs || len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	cfg := &packages.Config{
		Mode: packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedDeps |
			packages.NeedImports,
		Dir: ".",
		Env: os.Environ(),
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatalf("packages.Load: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		log.Fatal("type-check errors present — aborting")
	}

	// ── gather
	type ifacePair struct {
		named *types.Named
		iface *types.Interface
	}
	var ifaces []ifacePair
	var concretes []*types.Named

	for _, p := range pkgs {
		for _, name := range p.Types.Scope().Names() {
			if tn, ok := p.Types.Scope().Lookup(name).(*types.TypeName); ok {
				named := tn.Type().(*types.Named)
				switch u := named.Underlying().(type) {
				case *types.Interface:
					ifaces = append(ifaces, ifacePair{named, u.Complete()})
				default:
					concretes = append(concretes, named)
				}
			}
		}
	}

	sort.Slice(ifaces, func(i, j int) bool { return q(ifaces[i].named) < q(ifaces[j].named) })
	sort.Slice(concretes, func(i, j int) bool { return q(concretes[i]) < q(concretes[j]) })

	// ── check T and *T
	seen := map[string]struct{}{}
	for _, c := range concretes {
		val := c // val is types.Type
		ptr := types.NewPointer(val)

		for _, ip := range ifaces {
			if types.Implements(val, ip.iface) {
				key := q(c) + " → " + q(ip.named)
				seen[key] = struct{}{}
			}
			if types.Implements(ptr, ip.iface) {
				key := "*" + q(c) + " → " + q(ip.named)
				seen[key] = struct{}{}
			}
		}
	}

	// ── print in lexical order
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k)
	}
}

func q(n *types.Named) string {
	if n.Obj().Pkg() == nil {
		return n.Obj().Name()
	}
	return n.Obj().Pkg().Name() + "." + n.Obj().Name()
}
