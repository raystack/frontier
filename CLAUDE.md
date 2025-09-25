# GRPC to Connect RPC Migration Instructions

I am migrating grpc handlers to buf connect rpc handler. The handlers are in `internal/api/v1beta1/` and they are to be migrated to `internal/api/v1beta1connect` package.

## Handler Interface Changes
The handler should wrap the request and response types as follows:

**Request parameter change:**
- From: `req *frontierv1beta1.ListProjectsRequest`
- To: `req *connect.Request[frontierv1beta1.ListProjectsRequest]`

**Response return change:**
- From: `(*frontierv1beta1.ListProjectsResponse, error)`
- To: `(*connect.Response[frontierv1beta1.ListProjectsResponse], error)`

**Accessing request data:**
- Use `request.Msg.GetFieldName()` instead of `request.GetFieldName()`

**Creating responses:**
- Use `connect.NewResponse(&frontierv1beta1.ResponseType{...})` to wrap responses

## Logger Integration
**Always preserve logger functionality:**
- Add `grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"` import
- Extract logger at the beginning of each handler: `logger := grpczap.Extract(ctx)` if required
- Use logger in exactly the same places and manner as the original gRPC handler
- Add any missing error constants to `errors.go` if the original uses logger.Error() calls

## Error Handling
Transform all error handling to use Connect RPC error codes:

**Standard error mapping:**
- Use appropriate Connect error codes from `"connectrpc.com/connect"` package
- For unknown/internal errors: `connect.NewError(connect.CodeInternal, ErrInternalServerError)`
- For validation errors: `connect.NewError(connect.CodeInvalidArgument, err)`
- For not found errors: `connect.NewError(connect.CodeNotFound, ErrNotFound)` or `ErrUserNotExist`
- For conflicts: `connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)`

**Error constants:**
- Use the errors defined in `errors.go` in the v1beta1connect package
- If original gRPC handler uses specific error constants, add them to `errors.go`
- Maintain exact same error messages as gRPC version for consistency

## Constants and Schema Management
**Metaschema constants:**
- Add any required schema constants to `metaschema.go` (e.g., `userMetaSchema = "user"`)
- Check original gRPC handler for metaschema validation usage

**Error constants:**
- Add missing error constants to `errors.go` following the existing pattern
- Maintain exact error messages from gRPC handlers for consistency

## Testing Migration
When migrating tests from `internal/api/v1beta1/` to `internal/api/v1beta1connect/`:

**Test structure changes:**
- Use `*connect.Request[RequestType]` and `*connect.Response[ResponseType]` in test tables
- Create requests using `connect.NewRequest(&RequestType{...})`
- Create expected responses using `connect.NewResponse(&ResponseType{...})`
- Use `&ConnectHandler{}` instead of `Handler{}` for the handler instance

**Comprehensive test coverage:**
- Migrate ALL test cases from the original gRPC test
- Ensure coverage for all error scenarios (validation, not found, conflicts, internal errors)
- Include success scenarios with proper data transformation
- Test edge cases like empty strings, invalid UUIDs, etc.

**Required test imports:**
- Add `"github.com/raystack/frontier/pkg/utils"` if using `utils.NewString()` for random IDs
- Include imports for any context setup (e.g., authentication context)

**Error expectations:**
- Use Connect error types: `connect.NewError(connect.CodeInternal, ErrInternalServerError)`
- Match error codes to expected scenarios (NotFound, InvalidArgument, AlreadyExists, etc.)

## Migration Checklist
- [ ] Update function signature (request/response wrapping)
- [ ] Change request field access to use `request.Msg.GetFieldName()`
- [ ] Wrap response with `connect.NewResponse()`
- [ ] Convert all errors to Connect error format
- [ ] Add logger extraction: `logger := grpczap.Extract(ctx)`
- [ ] Add grpczap import: `grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"`
- [ ] Add any missing error constants to `errors.go`
- [ ] Add any required schema constants to `metaschema.go`
- [ ] Add all necessary imports based on original handler dependencies
- [ ] Migrate corresponding tests with comprehensive coverage
- [ ] Test with `make build` to ensure compilation
- [ ] Run specific tests to verify functionality
- [ ] Keep all business logic unchanged - only interface changes

## Best Practices & Common Patterns
**String handling:**
- Always use `strings.TrimSpace()` where the original gRPC handler does
- Preserve exact same validation logic and error conditions

**Context handling:**
- Use `authenticate.GetEmailFromContext(ctx)` for email retrieval from context
- Preserve all context-based operations exactly as in original

**Metadata handling:**
- Use `metadata.Build()` and metaschema validation where original does
- Add required metaschema constants to `metaschema.go`

**Business logic preservation:**
- Keep all business logic unchanged - only interface changes
- Maintain exact same error conditions and success scenarios
- Preserve all audit logging, user slug generation, and other operations

## Authentication and Principal-Based APIs
When migrating APIs that require authentication (current user operations):

**Principal authentication pattern:**
- Use `h.GetLoggedInPrincipal(ctx)` instead of `h.authnService.GetPrincipal()` directly
- Return the error directly from `GetLoggedInPrincipal()` without wrapping it - it already returns proper Connect error codes
- For current user operations, validate that request email matches principal user email where applicable

**Error propagation:**
- `GetLoggedInPrincipal()` returns proper Connect error codes (CodeUnauthenticated, CodeNotFound, etc.)
- Don't wrap authentication errors in `CodeInternal` - let them propagate as-is
- This ensures proper error codes reach the client (e.g., 401 Unauthenticated instead of 500 Internal)

## Advanced Business Logic Patterns
**Email upsert logic:**
- Some APIs like UpdateUser have complex business logic (update by email ID with create fallback)
- Preserve all conditional logic, fallback mechanisms, and edge case handling exactly
- Test all code paths including fallback scenarios

**Metadata validation:**
- Always preserve metadata schema validation where present in original
- Add required schema constants to `metaschema.go`
- Maintain exact same validation error messages

## Test Migration Best Practices
**Mock expectations:**
- Use `mock.Anything` for context parameters instead of specific context types
- Context types vary between test runs and can cause brittle tests
- For other parameters, use specific expectations or `mock.AnythingOfType()` as needed

**Test case completeness:**
- Always migrate ALL test cases from original gRPC tests - no exceptions
- Include edge cases, validation scenarios, success cases, and error conditions
- Ensure test coverage matches or exceeds original implementation

**Error code validation:**
- For success cases, check `err == nil` and validate response content
- For error cases, use `connect.CodeOf(err)` to check specific Connect error codes
- Use `connect.Code(0)` pattern for success case comparisons in test tables

**Early return validation:**
- When functions return early (e.g., empty request body), don't expect subsequent mock calls
- Structure test setup to only mock what will actually be called in that scenario

## Other Requirements
- Do not commit this CLAUDE.md file
- Do not commit anything in scripts folder
- Do not put Linear ticket reference in PR description or commit messages
- Always test the migration with `make build` and run tests to verify functionality