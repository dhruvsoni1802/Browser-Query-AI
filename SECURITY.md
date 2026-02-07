# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please follow these steps:

### 1. **Do NOT** Open a Public Issue

Security vulnerabilities should not be disclosed publicly until a fix is available.

### 2. Report Privately

**Email**: dhruv.soni@example.com (replace with your actual email)

**Subject**: `[SECURITY] Brief description of vulnerability`

Include in your report:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if you have one)
- Your contact information

### 3. Response Timeline

- **Within 48 hours**: We'll acknowledge receipt of your report
- **Within 7 days**: We'll provide a detailed response with next steps
- **Within 30 days**: We'll release a fix or workaround (if feasible)

### 4. Disclosure Policy

- We'll work with you to understand and resolve the issue
- Once a fix is available, we'll publicly disclose the vulnerability
- We'll credit you in the security advisory (unless you prefer anonymity)

## Security Considerations

### Browser Process Isolation

This project manages headless browser processes. Be aware of:

#### 1. **Sandbox Bypass**

Browser processes run with `--no-sandbox` flag for compatibility. This reduces security isolation.

**Mitigation:**
- Run the service in a containerized environment (Docker)
- Use separate OS-level users for browser processes
- Implement network-level restrictions

#### 2. **Arbitrary Code Execution**

The service executes JavaScript in browser contexts via CDP.

**Risks:**
- Malicious scripts could exploit browser vulnerabilities
- Untrusted URLs could lead to XSS or other attacks

**Mitigation:**
- Validate and sanitize all URLs before navigation
- Implement URL allowlist/blocklist
- Use isolated browser contexts per client
- Set resource limits (CPU, memory, network)

#### 3. **Resource Exhaustion**

Malicious actors could exhaust system resources.

**Risks:**
- Creating too many browser contexts
- Opening too many pages
- Infinite loops in JavaScript execution
- Memory leaks

**Mitigation:**
- Enforce limits on contexts per session
- Implement timeouts for all operations
- Monitor resource usage and implement circuit breakers
- Use port pool to limit concurrent browsers

### WebSocket Security

#### 1. **Connection Hijacking**

CDP WebSocket connections are unauthenticated by default.

**Risks:**
- Unauthorized access to browser debugging interface
- Ability to execute arbitrary commands

**Mitigation:**
- Bind debug ports to localhost only
- Use firewall rules to restrict access
- Implement authentication layer above CDP
- Rotate ports periodically

#### 2. **Command Injection**

CDP commands accept arbitrary parameters.

**Risks:**
- Malicious parameters could crash browsers
- Exploit browser vulnerabilities

**Mitigation:**
- Validate all command parameters
- Use type-safe wrappers around CDP commands
- Implement allowlist of permitted CDP methods

### Data Privacy

#### 1. **Cookie and Storage Leakage**

Browser contexts maintain cookies and local storage.

**Risks:**
- Session data persisting longer than intended
- Cross-client data leakage if contexts aren't properly isolated

**Mitigation:**
- Always dispose browser contexts after use
- Implement session timeouts
- Use separate contexts for different clients
- Clear browser data on context disposal

#### 2. **Screenshot Data**

Screenshots may contain sensitive information.

**Risks:**
- Credentials visible in screenshots
- PII (Personally Identifiable Information) exposure

**Mitigation:**
- Encrypt screenshot data in transit
- Don't log screenshot contents
- Implement retention policies
- Sanitize screenshots before storage

### Network Security

#### 1. **SSRF (Server-Side Request Forgery)**

The service navigates to user-provided URLs.

**Risks:**
- Access to internal network resources
- Port scanning
- Cloud metadata access (e.g., AWS metadata endpoint)

**Mitigation:**
- Implement URL validation and filtering
- Block private IP ranges (RFC 1918)
- Block cloud metadata IPs (169.254.169.254)
- Use DNS allowlists
- Implement request timeouts

**Example blocklist:**
```
10.0.0.0/8
172.16.0.0/12
192.168.0.0/16
127.0.0.0/8
169.254.169.254/32  # AWS metadata
```

#### 2. **DDoS Protection**

The service could be used to generate traffic to external sites.

**Risks:**
- Unintentional DDoS on target websites
- Reputation damage
- Legal liability

**Mitigation:**
- Rate limiting per client/session
- Request throttling
- Monitor outbound traffic
- Implement backpressure mechanisms

### Dependency Security

#### 1. **Supply Chain Attacks**

Go dependencies could contain vulnerabilities.

**Mitigation:**
- Regularly update dependencies: `go get -u ./...`
- Use `go mod verify` to check module integrity
- Review dependency changes in PRs
- Use GitHub Dependabot for automated updates
- Scan dependencies with `govulncheck`
```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for vulnerabilities
govulncheck ./...
```

#### 2. **Browser Vulnerabilities**

Chromium itself may have security issues.

**Mitigation:**
- Use latest stable Chrome/Chromium version
- Monitor Chrome security bulletins
- Update browser binaries regularly
- Consider using Chrome for Testing builds

### Configuration Security

#### 1. **Environment Variables**

Sensitive configuration should not be hardcoded.

**Best Practices:**
- Use environment variables for secrets
- Never commit `.env` files to version control
- Use secret management systems in production (e.g., HashiCorp Vault, AWS Secrets Manager)
- Implement least-privilege access

#### 2. **File Permissions**

Browser user data directories contain sensitive information.

**Best Practices:**
- Use `0600` permissions for sensitive files
- Run service with dedicated user (not root)
- Clean up temporary directories on shutdown
- Encrypt browser data directories if possible

## Security Best Practices for Deployment

### 1. **Container Security**

When deploying with Docker:
```dockerfile
# Use minimal base image
FROM golang:1.21-alpine AS builder

# Run as non-root user
RUN adduser -D -u 1000 appuser
USER appuser

# Drop unnecessary capabilities
# Use read-only root filesystem where possible
```

### 2. **Network Isolation**

- Use internal networks for browser processes
- Expose only necessary ports
- Implement TLS for external APIs
- Use reverse proxy (nginx, Traefik) with rate limiting

### 3. **Monitoring and Logging**

- Log all security-relevant events
- Monitor for suspicious patterns (rapid context creation, unusual URLs)
- Implement alerting for security events
- Rotate and secure logs

### 4. **Resource Limits**
```yaml
# Docker Compose example
services:
  browser-query-ai:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 512M
```

## Known Security Limitations

### Current Version (0.1.x)

1. **No Authentication**: The service does not implement authentication. Deploy behind an authenticated proxy in production.

2. **No Rate Limiting**: No built-in rate limiting. Implement at the infrastructure level.

3. **No URL Filtering**: All URLs are permitted by default. Implement allowlist/blocklist as needed.

4. **No Encryption at Rest**: Browser data is not encrypted on disk.

5. **Shared Browser Processes**: Multiple contexts share the same browser process (resource optimization vs. isolation trade-off).

## Security Roadmap

Planned security improvements:

- [ ] Implement authentication layer
- [ ] Add URL filtering and validation
- [ ] Implement rate limiting and circuit breakers
- [ ] Add encryption for browser data directories
- [ ] Implement audit logging
- [ ] Add SSRF protection
- [ ] Security scanning in CI/CD pipeline
- [ ] Regular security audits

## Security Champions

If you're interested in helping improve security, we'd love your input! Join our security discussions or reach out.

## Acknowledgments

We appreciate responsible disclosure and will acknowledge security researchers who report vulnerabilities:

- **Hall of Fame**: Security researchers who have helped improve this project
  - (None yet - be the first!)

---

**Remember**: Security is everyone's responsibility. If you see something, say something.