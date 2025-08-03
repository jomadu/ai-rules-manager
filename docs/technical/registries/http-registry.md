# HTTP Registry Guide

## Overview

ARM's Generic HTTP registry provides simple package hosting using any HTTP file server. It focuses on direct downloads with predictable URL patterns, making it ideal for basic setups without complex infrastructure.

## Features

- **Simple Setup**: Works with any HTTP file server
- **Direct Downloads**: Predictable URL patterns for package access
- **Authentication**: Bearer token or basic auth support
- **No Version Discovery**: Requires exact version specification
- **Minimal Infrastructure**: No special server requirements

## Directory Structure

### Basic Structure
```
https://registry.example.com/
├── typescript-rules/
│   ├── 1.0.0.tar.gz
│   ├── 1.0.1.tar.gz
│   └── 1.1.0.tar.gz
└── security-rules/
    ├── 2.0.0.tar.gz
    └── 2.1.0.tar.gz
```

### With Organization Scope
```
https://registry.example.com/
├── company/
│   ├── typescript-rules/
│   │   ├── 1.0.0.tar.gz
│   │   └── 1.0.1.tar.gz
│   └── security-rules/
│       └── 1.0.0.tar.gz
└── opensource/
    └── lint-rules/
        └── 0.1.0.tar.gz
```

## Configuration Examples

### Basic HTTP Registry
```ini
[sources.http]
type = generic
url = https://registry.example.com
```

### HTTP Registry with Authentication
```ini
[sources.company-http]
type = generic
url = https://internal.company.com/packages
authToken = ${COMPANY_REGISTRY_TOKEN}
```

### HTTP Registry with Basic Auth
```ini
[sources.private-http]
type = generic
url = https://private-registry.com
authToken = Basic ${BASE64_CREDENTIALS}
```

## URL Patterns

### Download URLs
- **No org**: `https://registry.example.com/pkg/version.tar.gz`
- **With org**: `https://registry.example.com/org/pkg/version.tar.gz`

### Examples
```
# Direct package downloads
https://registry.example.com/typescript-rules/1.0.0.tar.gz
https://registry.example.com/company/security-rules/2.0.0.tar.gz
```

## Usage Examples

### Install Commands
```bash
# Must specify exact version - no discovery
arm install typescript-rules@1.0.0
arm install company@security-rules@2.0.0

# This will fail - no version discovery
arm install typescript-rules  # Error: version required
```

### Configuration Usage
```bash
# Configure HTTP registry
arm config set sources.company https://internal.company.com/packages
arm config set sources.company.authToken ${TOKEN}

# Install from configured source
arm install company@typescript-rules@1.0.0
```

## Publishing Workflow

### Manual Upload
```bash
# Upload to web server directory
scp typescript-rules-1.0.0.tar.gz user@server:/var/www/registry/typescript-rules/
```

### Automated CI/CD
```yaml
# GitHub Actions example
- name: Upload to HTTP Registry
  run: |
    curl -X PUT \
      -H "Authorization: Bearer ${{ secrets.REGISTRY_TOKEN }}" \
      --data-binary @${PACKAGE_NAME}-${VERSION}.tar.gz \
      https://registry.example.com/${PACKAGE_NAME}/${VERSION}.tar.gz
```

### With Organization Scope
```bash
# Upload scoped package
scp security-rules-1.0.0.tar.gz user@server:/var/www/registry/company/security-rules/
```

## Server Setup Examples

### Nginx Configuration
```nginx
server {
    listen 80;
    server_name registry.example.com;
    root /var/www/registry;

    location / {
        try_files $uri $uri/ =404;
        add_header Access-Control-Allow-Origin *;
    }

    # Optional: Basic authentication
    location /private/ {
        auth_basic "Registry Access";
        auth_basic_user_file /etc/nginx/.htpasswd;
    }
}
```

### Apache Configuration
```apache
<VirtualHost *:80>
    ServerName registry.example.com
    DocumentRoot /var/www/registry

    <Directory /var/www/registry>
        Options Indexes FollowSymLinks
        AllowOverride None
        Require all granted
    </Directory>

    # Optional: Basic authentication
    <Directory /var/www/registry/private>
        AuthType Basic
        AuthName "Registry Access"
        AuthUserFile /etc/apache2/.htpasswd
        Require valid-user
    </Directory>
</VirtualHost>
```

### Simple Python Server
```python
# Simple development server
import http.server
import socketserver

PORT = 8000
Handler = http.server.SimpleHTTPRequestHandler

with socketserver.TCPServer(("", PORT), Handler) as httpd:
    print(f"Serving at port {PORT}")
    httpd.serve_forever()
```

## Best Practices

### Server Configuration
- Enable HTTPS for production deployments
- Configure proper CORS headers for web access
- Set up appropriate caching headers
- Use CDN for global distribution

### Authentication
- Use bearer tokens for API-style access
- Implement basic auth for simple setups
- Consider IP-based restrictions for internal registries
- Rotate authentication tokens regularly

### Organization
- Use consistent URL patterns
- Organize packages by organization/team
- Implement proper directory permissions
- Monitor server access logs

### Performance
- Enable gzip compression for faster downloads
- Use CDN for frequently accessed packages
- Implement proper caching strategies
- Monitor bandwidth usage

## Limitations

### No Version Discovery
HTTP registries cannot list available versions without directory listing support:
```bash
# This will fail
arm install typescript-rules

# Must specify exact version
arm install typescript-rules@1.0.0
```

### No Rich Metadata
HTTP registries provide minimal metadata:
- Package name and repository URL only
- No version information, download counts, or descriptions
- No dependency information

### Server Dependencies
- Requires HTTP server setup and maintenance
- No built-in authentication or access control
- Manual package upload and organization

## Troubleshooting

### Common Issues
1. **404 Not Found**: Check URL pattern and file placement
2. **403 Forbidden**: Verify authentication credentials
3. **CORS Errors**: Configure proper CORS headers
4. **Version Required**: HTTP registries need exact versions

### Debugging Steps
```bash
# Test direct download
curl -I https://registry.example.com/typescript-rules/1.0.0.tar.gz

# Test with authentication
curl -H "Authorization: Bearer $TOKEN" \
  https://registry.example.com/company/package/1.0.0.tar.gz

# Check ARM configuration
arm config get sources.company
```

### Server Logs
Monitor HTTP server logs for:
- 404 errors indicating missing packages
- 403 errors indicating authentication issues
- High bandwidth usage patterns
- Unusual access patterns
