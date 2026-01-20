package sqlbuilders

import (
	"go/ast"
	"go/token"
	"strings"
)

// SQLCChecker checks for SELECT * in sqlc generated code.
type SQLCChecker struct{}

// NewSQLCChecker creates a new sqlc checker.
func NewSQLCChecker() *SQLCChecker {
	return &SQLCChecker{}
}

// Name returns the checker name.
func (c *SQLCChecker) Name() string {
	return "sqlc"
}

// IsApplicable checks if the call is from sqlc generated code.
func (c *SQLCChecker) IsApplicable(call *ast.CallExpr) bool {
	// sqlc generates code with patterns like:
	// - Queries struct methods
	// - db.QueryRow or db.Query with specific patterns

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check for common sqlc patterns
	methodName := sel.Sel.Name
	sqlcMethods := []string{
		"GetUser", "ListUsers", "CreateUser", "UpdateUser", "DeleteUser",
		"Get", "List", "Create", "Update", "Delete", "Find", "Search",
	}

	for _, m := range sqlcMethods {
		if strings.Contains(methodName, m) {
			return true
		}
	}

	return false
}

// CheckSelectStar checks for SELECT * in the call.
func (c *SQLCChecker) CheckSelectStar(call *ast.CallExpr) *SelectStarViolation {
	// sqlc doesn't typically have SELECT * visible in Go code
	// but we can check string arguments
	for _, arg := range call.Args {
		if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			value := strings.ToUpper(lit.Value)
			if strings.Contains(value, "SELECT *") || strings.Contains(value, "SELECT\t*") {
				return &SelectStarViolation{
					Pos:     lit.Pos(),
					End:     lit.End(),
					Message: "sqlc query contains SELECT * - specify columns explicitly in your .sql file",
				}
			}
		}
	}
	return nil
}

// CheckChainedCalls checks chained method calls.
func (c *SQLCChecker) CheckChainedCalls(call *ast.CallExpr) []*SelectStarViolation {
	// sqlc doesn't typically use chained calls
	return nil
}
