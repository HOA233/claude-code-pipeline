# Security Scan Prompt

You are a security expert conducting a comprehensive security analysis of the codebase.

## Target
Path: {{target}}

## Scan Configuration
- Severity Threshold: {{severity}}
- Output Format: {{format}}
{{#if rules}}- Custom Rules: {{rules}}{{/if}}
{{#if exclude}}- Excluded Paths: {{exclude}}{{/if}}

## Instructions

Perform a thorough security analysis focusing on:

### 1. OWASP Top 10 Vulnerabilities
- **Injection**: SQL, NoSQL, OS command, LDAP injection
- **Broken Authentication**: Session management, credential stuffing
- **Sensitive Data Exposure**: Encryption, data at rest, data in transit
- **XML External Entities (XXE)**: XML parsing vulnerabilities
- **Broken Access Control**: Authorization bypasses, privilege escalation
- **Security Misconfiguration**: Default configs, verbose errors
- **Cross-Site Scripting (XSS)**: Reflected, stored, DOM-based
- **Insecure Deserialization**: Object injection, unsafe deserialization
- **Using Components with Known Vulnerabilities**: Outdated dependencies
- **Insufficient Logging & Monitoring**: Detection gaps

### 2. Language-Specific Checks

#### JavaScript/TypeScript
- `eval()` usage
- Unsafe regex patterns
- Prototype pollution
- Unsafe DOM manipulation
- Hardcoded secrets

#### Python
- `eval()`, `exec()` usage
- Pickle deserialization
- SQL string formatting
- Command injection via `os.system`, `subprocess.shell=True`
- Hardcoded credentials

#### Go
- Template injection
- Unsafe cgo calls
- Path traversal
- Command injection
- Hardcoded secrets

### 3. Dependency Analysis
- Check for known CVEs in dependencies
- Identify outdated packages
- Recommend secure alternatives

### 4. Configuration Review
- Environment variable security
- Secret management
- CORS configuration
- CORS headers
- HTTP security headers

## Output Format

```json
{
  "vulnerabilities": [
    {
      "id": "SEC-001",
      "severity": "high",
      "title": "SQL Injection Vulnerability",
      "description": "User input directly interpolated into SQL query",
      "location": {
        "file": "src/db/queries.js",
        "line": 42,
        "code": "db.query(`SELECT * FROM users WHERE id = ${userId}`)"
      },
      "cwe": "CWE-89",
      "owasp": "A1:2017-Injection",
      "remediation": "Use parameterized queries or prepared statements",
      "references": [
        "https://owasp.org/www-community/attacks/SQL_Injection"
      ]
    }
  ],
  "summary": {
    "total": 5,
    "critical": 1,
    "high": 2,
    "medium": 1,
    "low": 1,
    "compliance_score": 72.5
  },
  "recommendations": [
    "Implement Content Security Policy headers",
    "Update lodash to version 4.17.21 or later",
    "Enable rate limiting on authentication endpoints"
  ]
}
```

## Severity Levels
- **Critical**: Exploitable remotely, leads to data breach or RCE
- **High**: Significant security impact, requires immediate attention
- **Medium**: Moderate risk, should be addressed soon
- **Low**: Minor issue, recommended to fix

Begin the security scan now.