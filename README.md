[![Go Test Status](https://github.com/SunMaungOo/task-runner-auth/actions/workflows/go-test-ci.yaml/badge.svg)](https://github.com/SunMaungOo/task-runner-auth/actions/workflows/go-test-ci.yaml/badge.svg)


# Overview

Authentication service for task-runner.

Service that registry with email and password and return the JWT token if login success.

# Endpoint

| Method | Endpoint |
| ------------- |:-------------:|
| GET      |  /healthz     |
| GET      |  /readyz    |
| POST     | /register     |
| POST     | /login     |