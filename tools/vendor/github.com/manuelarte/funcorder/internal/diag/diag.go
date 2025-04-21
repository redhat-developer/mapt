package diag

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func NewConstructorNotAfterStructType(structSpec *ast.TypeSpec, constructor *ast.FuncDecl) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: constructor.Pos(),
		Message: fmt.Sprintf("function %q for struct %q should be placed after the struct declaration",
			constructor.Name, structSpec.Name),
	}
}

func NewConstructorNotBeforeStructMethod(
	structSpec *ast.TypeSpec,
	constructor *ast.FuncDecl,
	method *ast.FuncDecl,
) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: constructor.Pos(),
		Message: fmt.Sprintf("constructor %q for struct %q should be placed before struct method %q",
			constructor.Name, structSpec.Name, method.Name),
	}
}

func NewAdjacentConstructorsNotSortedAlphabetically(
	structSpec *ast.TypeSpec,
	constructorNotSorted *ast.FuncDecl,
	otherConstructorNotSorted *ast.FuncDecl,
) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: otherConstructorNotSorted.Pos(),
		Message: fmt.Sprintf("constructor %q for struct %q should be placed before constructor %q",
			otherConstructorNotSorted.Name, structSpec.Name, constructorNotSorted.Name),
	}
}

func NewNonExportedMethodBeforeExportedForStruct(
	structSpec *ast.TypeSpec,
	privateMethod *ast.FuncDecl,
	publicMethod *ast.FuncDecl,
) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: privateMethod.Pos(),
		Message: fmt.Sprintf("unexported method %q for struct %q should be placed after the exported method %q",
			privateMethod.Name, structSpec.Name, publicMethod.Name),
	}
}

func NewAdjacentStructMethodsNotSortedAlphabetically(
	structSpec *ast.TypeSpec,
	method *ast.FuncDecl,
	otherMethod *ast.FuncDecl,
) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: otherMethod.Pos(),
		Message: fmt.Sprintf("method %q for struct %q should be placed before method %q",
			otherMethod.Name, structSpec.Name, method.Name),
	}
}
