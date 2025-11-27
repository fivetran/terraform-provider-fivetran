# Terraform Provider Implementation Estimate
## Connector SDK Terraform Provider Support - Terraform Side

Based on Technical Design Document and codebase analysis, this document provides detailed time estimates for the Terraform provider implementation, testing, and deployment.

---

## Work Breakdown Structure

### Phase 2: Terraform Provider Implementation

#### 2.1 Go-Fivetran Client Implementation
**Files to Create/Modify:**
- `go-fivetran/connector_sdk_deployment/connector_sdk_deployment_create.go`
- `go-fivetran/connector_sdk_deployment/connector_sdk_deployment_details.go`
- `go-fivetran/connector_sdk_deployment/connector_sdk_deployment_delete.go`
- `go-fivetran/connector_sdk_deployment/common_types.go`
- `go-fivetran/connector_sdk_deployment/connector_sdk_deployment_create_test.go`
- `go-fivetran/connector_sdk_deployment/connector_sdk_deployment_details_test.go`
- `go-fivetran/connector_sdk_deployment/connector_sdk_deployment_delete_test.go`
- `go-fivetran/fivetran.go` (register new service)

**Complexity Notes:**
- Multipart form data upload handling (new pattern in go-fivetran)
- Base64 decoding of package content
- File upload with proper content-type headers
- Error handling for file size limits and validation

**Estimated Time: 5 developer days**
- Create service methods: 2 days
- Multipart upload implementation: 1.5 days
- Unit tests: 1 day
- Code review and refinement: 0.5 days

---

#### 2.2 Terraform Provider Resource: connector_sdk_deployment
**Files to Create:**
- `fivetran/framework/resources/connector_sdk_deployment.go`
- `fivetran/framework/core/schema/connector_sdk_deployment.go`
- `fivetran/framework/core/model/connector_sdk_deployment.go`

**Complexity Notes:**
- Base64 decoding of `package_name` input
- File validation (size limits, base64 format)
- Computed attributes (id, created_at, updated_at)
- Import state functionality
- Full CRUD operations
- State management for file uploads (prevent unnecessary re-uploads)

**Estimated Time: 4 developer days**
- Resource implementation: 2 days
- Schema and model: 0.5 days
- Validation logic: 0.5 days
- Import functionality: 0.5 days
- Code review and refinement: 0.5 days

---

#### 2.3 Extend Connector Resource
**Files to Modify:**
- `fivetran/framework/resources/connector.go`
- `fivetran/framework/core/schema/connector.go`
- `fivetran/framework/core/model/connector.go`
- `fivetran/framework/core/model/connector_resource.go`

**New Fields to Add:**
- `deployment_id` (reference to deployment resource)
- `python_version` (optional, validation required)
- `connector_files` (optional, file name)
- `secrets_list` (optional, JSON-encoded secrets)

**Complexity Notes:**
- Integration with existing connector config structure
- Conditional validation (only for service = "connector_sdk")
- Dependency handling (deployment_id must exist)
- Backward compatibility maintenance
- Secrets handling (sensitive data)

**Estimated Time: 3 developer days**
- Schema extension: 0.5 days
- Model extension: 0.5 days
- Resource logic updates: 1 day
- Validation logic: 0.5 days
- Testing integration: 0.5 days

---

#### 2.4 Testing: Unit and Integration Tests

**Files to Create:**
- `fivetran/tests/mock/resource_connector_sdk_deployment_test.go`
- `fivetran/tests/e2e/resource_connector_sdk_deployment_e2e_test.go`
- `fivetran/tests/e2e/resource_connector_with_sdk_deployment_e2e_test.go`
- `go-fivetran/connector_sdk_deployment/*_test.go` (already counted in 2.1)

**Test Coverage:**
- Deployment resource CRUD operations
- Base64 encoding/decoding validation
- File upload error handling
- Deployment deletion and cleanup
- Connector with deployment_id reference
- Import functionality
- State management
- Error scenarios (invalid base64, file too large, etc.)

**Estimated Time: 4 developer days**
- Mock tests: 1.5 days
- E2E tests: 2 days
- Test data and fixtures: 0.5 days

---

#### 2.5 Documentation

**Files to Create/Modify:**
- `docs/resources/connector_sdk_deployment.md` (auto-generated from schema)
- `docs/resources/connector.md` (update with new fields)
- `config-examples/connector_sdk_deployment_example.tf` (new)
- `config-examples/connector_with_sdk_deployment_example.tf` (new)
- `CHANGELOG.md` (update)

**Documentation Tasks:**
- Resource documentation (auto-generated)
- Usage examples
- Configuration examples
- Migration guide (if needed)
- Changelog entries

**Estimated Time: 1.5 developer days**
- Example configurations: 0.5 days
- Documentation review and updates: 0.5 days
- Changelog and release notes: 0.5 days

---

#### 2.6 Provider Release Preparation

**Tasks:**
- Version bump
- Changelog finalization
- Provider build and validation
- Release checklist completion
- Integration testing with backend API (coordination with backend team)

**Estimated Time: 1 developer day**
- Build validation: 0.25 days
- Release preparation: 0.25 days
- Final integration testing: 0.5 days

---

## Summary Table

| Work Type | Role | Short Description | Developer Days |
|-----------|------|-------------------|----------------|
| Implement | SWE | Go-fivetran client implementation (multipart upload, CRUD) | 5 |
| Implement | SWE | Terraform deployment resource (CRUD, validation, import) | 4 |
| Implement | SWE | Extend connector resource (new fields, validation) | 3 |
| QA | SWE/SDET | Unit tests, integration tests, E2E tests | 4 |
| Release | SWE | Documentation, examples, changelog, release prep | 2.5 |
| **Total** | | | **18.5** |

---

## Risk Factors and Contingency

### High Risk Areas:
1. **Multipart File Upload Implementation** (+2 days contingency)
   - First multipart upload implementation in go-fivetran
   - May require http_utils extension
   - Testing with large files

2. **Base64 Decoding and File Handling** (+1 day contingency)
   - Memory management for large files
   - Validation edge cases
   - Error handling

3. **Backend API Coordination** (+1 day contingency)
   - API endpoint availability timing
   - Response format changes
   - Integration testing dependencies

### Medium Risk Areas:
1. **Connector Resource Extension** (+0.5 days contingency)
   - Backward compatibility testing
   - State migration (if needed)
   - Conditional validation logic

2. **E2E Testing** (+1 day contingency)
   - Test environment setup
   - Backend API availability
   - File upload test data preparation

### Contingency Buffer: **5.5 days**

---

## Adjusted Estimate with Contingency

| Work Type | Base Estimate | Contingency | Adjusted Estimate |
|-----------|---------------|-------------|-------------------|
| Go-fivetran Client | 5 | +2 | 7 |
| Deployment Resource | 4 | +1 | 5 |
| Connector Extension | 3 | +0.5 | 3.5 |
| Testing | 4 | +1 | 5 |
| Documentation & Release | 2.5 | +1 | 3.5 |
| **Total** | **18.5** | **+5.5** | **24** |

---

## Dependencies and Prerequisites

### External Dependencies:
1. **Backend API Endpoints** (Phase 1)
   - POST /v1/connector-sdk/deployments
   - GET /v1/connector-sdk/deployments/{deploymentId}
   - DELETE /v1/connector-sdk/deployments/{deploymentId}
   - Must be available before Terraform provider E2E testing

2. **Database Schema** (Phase 1)
   - connector_sdk_deployments table
   - Must be deployed before integration testing

3. **Backend Integration**
   - ConnectorSdkCredentials.deployment_id field
   - Connector creation/update with deployment_id
   - Must be available before connector resource extension testing

### Internal Dependencies:
1. **go-fivetran Client**
   - Must be updated before Terraform provider implementation
   - Version pinning and dependency management

2. **Test Environment**
   - Backend API test environment
   - File storage (GCS) test bucket
   - Test credentials and permissions

---

## Testing Strategy Breakdown

### Unit Tests (Mock Tests)
- Deployment resource CRUD operations: 8 test cases
- Base64 validation: 5 test cases
- File upload error handling: 5 test cases
- Connector with deployment: 5 test cases
- **Total: ~23 test cases, 1.5 days**

### Integration Tests (E2E)
- Full deployment lifecycle: 3 test cases
- Connector with deployment: 2 test cases
- Import functionality: 2 test cases
- Error scenarios: 3 test cases
- **Total: ~10 test cases, 2 days**

### Test Data Preparation
- Sample connector package (zip file)
- Base64 encoded test data
- Test fixtures and helpers
- **Total: 0.5 days**

---

## Timeline Considerations

### Sequential Dependencies:
1. **Week 1-2: Go-fivetran Client** (must complete before provider implementation)
2. **Week 2-3: Deployment Resource** (can start in parallel with client, but needs client for testing)
3. **Week 3: Connector Extension** (depends on deployment resource)
4. **Week 3-4: Testing** (depends on all implementations)
5. **Week 4: Documentation & Release** (depends on testing completion)

### Parallel Work Opportunities:
- Documentation can start early (schema-based)
- Example configurations can be prepared in parallel
- Test data preparation can happen in parallel with implementation

---

## Quality Assurance Considerations

### Code Review Requirements:
- All code must be reviewed by at least one other engineer
- Multipart upload implementation requires security review
- File handling requires performance review for large files

### Testing Requirements:
- 100% code coverage for new resource
- E2E tests must pass in test environment
- Backward compatibility must be verified
- Performance testing for large file uploads (>10MB)

### Documentation Requirements:
- All new resources must have usage examples
- Configuration examples must be tested
- Changelog must be comprehensive

---

## Notes and Assumptions

### Assumptions:
1. Backend API will be available according to Phase 1 timeline
2. Backend API response formats match TDD specifications
3. File upload size limits are consistent with existing Connector SDK limits
4. Test environment will be available for E2E testing
5. No breaking changes to existing Terraform provider patterns

### Unknowns:
1. Exact multipart upload implementation requirements (may need http_utils extension)
2. File size limits and validation requirements (need backend confirmation)
3. Error message formats from backend API
4. Deployment metadata response structure (final format)

### Mitigation:
- Regular sync with backend team during implementation
- Early API contract review
- Mock backend responses for early testing
- Incremental integration testing

---

## Sign-off

**Estimated by:** Jovan Manojlović  
**Date:** 2025-01-XX  
**Review Status:** Pending  
**Approval Status:** Pending

