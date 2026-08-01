# Internal Service Contracts

This directory contains versioned contracts used between GroundedSearch
components.

## Current contract

- `ml-service-v1.yaml`: health, readiness, and model metadata exposed by the
  Python ML service to the Go search API.

Breaking changes require a new contract version rather than silently changing
the existing file.
