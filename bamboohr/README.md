# Overview

This is a reference Pomerium data provider that returns employee data from BambooHR.
It has to be configured to connect to your BambooHR account, and would provide Pomerium
with subset of information that may be used to configure access policies.

# Use Cases

## Limit Access By Employee Properties

There may be business rules that require that employees access to certain resources
may be limited:

- `status` to deactivate access if an employee is terminated
- `department`, `division`, `jobTitle`
- `country`, `state`, `city`, `location`, `nationality`

# Configuration

For security reasons, all requests to BambooHR are runtime parameters
and may not be customized.

## Restricted fields

Fields `national_id, sin, ssn, nin` are restricted and may not be exported for security reasons.

# API

A `Token` HTTP header must be provided in each request, matching `api-token` runtime parameter.

## `GET /employees`

Returns list of current employees, with fields provided, as JSON array.

# `GET /employees/out`

Returns list of employees that are currently out.
