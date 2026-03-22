# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the Claude CLI Pipeline Service.

## What is an ADR?

An ADR is a document that captures an important architectural decision made along with its context and consequences.

## ADR Index

| Number | Title | Status |
|--------|-------|--------|
| 001 | Use Go as primary language | Accepted |
| 002 | Use Redis for state management | Accepted |
| 003 | Use Gin for HTTP framework | Accepted |
| 004 | CLI subprocess execution model | Accepted |
| 005 | Skill-based task execution | Accepted |

## Creating a new ADR

1. Copy `template.md` to a new file named `NNNN-title-with-dashes.md`
2. Fill in the sections
3. Submit as part of a pull request