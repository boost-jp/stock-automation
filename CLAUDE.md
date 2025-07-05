# Claude Development Notes

## Project Rules & Conventions

### R2 Domain Layer Implementation Rules

1. **Testing Framework**: Use `github.com/google/go-cmp/cmp` instead of `assert` for test comparisons
   - Use `cmp.Diff()` for detailed comparison output
   - Format error messages as: `(-want +got):\n%s`
   - Example: `if diff := cmp.Diff(expected, actual); diff != "" { t.Errorf("mismatch (-want +got):\n%s", diff) }`

2. **Domain Layer Structure**: 
   - Use SQLBoiler-generated models from `app/domain/models/`
   - Create domain services in `app/domain/` that use these models
   - Services should be stateless with constructor functions like `NewPortfolioService()`

3. **Type Conversions**:
   - Use simplified decimal conversions for now: `types.NullDecimal{}`
   - Domain services handle `types.Decimal` to `float64` conversions
   - Placeholder implementations are acceptable for complex decimal operations

4. **Test Coverage**:
   - Create comprehensive test files for each service: `*_service_test.go`
   - Test service creation, core business logic, validation, edge cases
   - Use table-driven tests for multiple scenarios
   - Include helper functions for test data creation

5. **Code Quality Before CI**:
   - Run `go mod tidy` to update dependencies
   - Run `go fmt` to format code
   - Run `go build` or `go test -c` to verify compilation
   - Add required dependencies to go.mod (like go-cmp)

6. **Import Management**:
   - Replace `interface{}` with `any` for modern Go
   - Remove unused imports (like cmpopts if not used)
   - Add necessary imports for testing dependencies

## Development Workflow

1. Create domain services from existing analysis logic
2. Write comprehensive tests using cmp.Diff
3. Update go.mod with required dependencies
4. Verify local compilation before creating PR
5. Document new patterns and rules in CLAUDE.md

## SQLBoiler Integration

- Generated models are in `app/domain/models/`
- Use `types.Decimal` and `types.NullDecimal` for price fields
- Convert between domain types and SQLBoiler types as needed
- Services act as adapters between domain logic and generated models