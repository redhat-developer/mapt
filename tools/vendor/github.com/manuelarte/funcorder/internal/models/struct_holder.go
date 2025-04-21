package models

import (
	"cmp"
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/manuelarte/funcorder/internal/diag"
	"github.com/manuelarte/funcorder/internal/features"
)

// StructHolder contains all the information around a Go struct.
type StructHolder struct {
	// The features to be analyzed
	Features features.Feature

	// The struct declaration
	Struct *ast.TypeSpec

	// A Struct constructor is considered if starts with `New...` and the 1st output parameter is a struct
	Constructors []*ast.FuncDecl

	// Struct methods
	StructMethods []*ast.FuncDecl
}

func (sh *StructHolder) AddConstructor(fn *ast.FuncDecl) {
	sh.Constructors = append(sh.Constructors, fn)
}

func (sh *StructHolder) AddMethod(fn *ast.FuncDecl) {
	sh.StructMethods = append(sh.StructMethods, fn)
}

// Analyze applies the linter to the struct holder.
func (sh *StructHolder) Analyze() []analysis.Diagnostic {
	// TODO maybe sort constructors and then report also, like NewXXX before MustXXX

	slices.SortFunc(sh.StructMethods, func(a, b *ast.FuncDecl) int {
		return cmp.Compare(a.Pos(), b.Pos())
	})

	var reports []analysis.Diagnostic

	if sh.Features.IsEnabled(features.ConstructorCheck) {
		reports = append(reports, sh.analyzeConstructor()...)
	}

	if sh.Features.IsEnabled(features.StructMethodCheck) {
		reports = append(reports, sh.analyzeStructMethod()...)
	}

	// TODO also check that the methods are declared after the struct
	return reports
}

func (sh *StructHolder) analyzeConstructor() []analysis.Diagnostic {
	var reports []analysis.Diagnostic

	for i, constructor := range sh.Constructors {
		if constructor.Pos() < sh.Struct.Pos() {
			reports = append(reports, diag.NewConstructorNotAfterStructType(sh.Struct, constructor))
		}

		if len(sh.StructMethods) > 0 && constructor.Pos() > sh.StructMethods[0].Pos() {
			reports = append(reports, diag.NewConstructorNotBeforeStructMethod(sh.Struct, constructor, sh.StructMethods[0]))
		}

		if sh.Features.IsEnabled(features.AlphabeticalCheck) &&
			i < len(sh.Constructors)-1 && sh.Constructors[i].Name.Name > sh.Constructors[i+1].Name.Name {
			reports = append(reports,
				diag.NewAdjacentConstructorsNotSortedAlphabetically(sh.Struct, sh.Constructors[i], sh.Constructors[i+1]),
			)
		}
	}
	return reports
}

func (sh *StructHolder) analyzeStructMethod() []analysis.Diagnostic {
	var lastExportedMethod *ast.FuncDecl

	for _, m := range sh.StructMethods {
		if !m.Name.IsExported() {
			continue
		}

		if lastExportedMethod == nil {
			lastExportedMethod = m
		}

		if lastExportedMethod.Pos() < m.Pos() {
			lastExportedMethod = m
		}
	}

	var reports []analysis.Diagnostic

	if lastExportedMethod != nil {
		for _, m := range sh.StructMethods {
			if m.Name.IsExported() || m.Pos() >= lastExportedMethod.Pos() {
				continue
			}

			reports = append(reports, diag.NewNonExportedMethodBeforeExportedForStruct(sh.Struct, m, lastExportedMethod))
		}
	}

	if sh.Features.IsEnabled(features.AlphabeticalCheck) {
		return slices.Concat(reports,
			isSorted(sh.Struct, filterMethods(sh.StructMethods, true)),
			isSorted(sh.Struct, filterMethods(sh.StructMethods, false)),
		)
	}

	return reports
}

func filterMethods(funcDecls []*ast.FuncDecl, exported bool) []*ast.FuncDecl {
	var result []*ast.FuncDecl

	for _, f := range funcDecls {
		if f.Name.IsExported() != exported {
			continue
		}

		result = append(result, f)
	}

	return result
}

func isSorted(typeSpec *ast.TypeSpec, funcDecls []*ast.FuncDecl) []analysis.Diagnostic {
	var reports []analysis.Diagnostic

	for i := range funcDecls {
		if i >= len(funcDecls)-1 {
			continue
		}

		if funcDecls[i].Name.Name > funcDecls[i+1].Name.Name {
			reports = append(reports,
				diag.NewAdjacentStructMethodsNotSortedAlphabetically(typeSpec, funcDecls[i], funcDecls[i+1]))
		}
	}
	return reports
}
