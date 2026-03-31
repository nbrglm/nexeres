# Rules for writing API contracts

This document outlines the rules and guidelines for writing contracts in our codebase. Adhering to these rules ensures consistency, readability, and maintainability across all contracts.

## General Guidelines
- NO METHODS OR LOGIC: Contracts should only define data structures and types. They must not contain any methods or business logic.
- USE TYPES: Always use defined types for properties instead of primitive types where applicable. This enhances clarity and consistency.
- OPTIONAL PROPERTIES: Optional properties must have a clear semantic meaning and documented behavior.
- PATCH SEMANTICS:
  - PATCH request objects must contain only optional (pointer) properties.
  - `nil` means no change.
  - Non-nil values apply updates.
  - Empty values explicitly clear fields.
  - PATCH must never be used for creation.
- NAMING CONVENTIONS: Follow the established naming conventions for types and properties.
  - Use PascalCase for type and property names (e.g., `UserProfile`, `FirstName`).
  - Use camelCase for property names in JSON (e.g., `json:"firstName"`).
- DOCUMENTATION: Provide clear and concise documentation for each contract, including descriptions for types and properties.
- DO NOT REFERENCE EXTERNAL LIBRARIES: Contracts should not reference types or structures from external libraries. They must be self-contained.
- For non-PATCH operations, empty values for optional properties must be explicitly allowed by validation rules; otherwise they are rejected.

## Pointer and Optionality Rules

- REQUIRED PROPERTIES:
  - Must be non-pointer types.
  - Must be present in the request.
  - Zero / empty values are invalid and must be rejected by validation.

- OPTIONAL PROPERTIES:
  - Must be pointer types.
  - `nil` means the property was not provided.
  - Non-nil means the property was provided and must be non-empty unless explicitly allowed.
  - For PATCH operations:
    - Non-nil + empty value means the property should be cleared.

## Forbidden Patterns

- Pointer to pointer types are not allowed.
- Pointer types must not be used for identifiers, ownership fields, or authentication context.
- Maps with unconstrained keys are disallowed in contracts.

## Enforcement

- Contracts violating these rules must fail validation at build time.
- SDK generation must fail if unsupported patterns are detected.
- Code reviews must ensure adherence to these rules before merging any changes involving contracts.