# HTTP Registry

Use generic HTTP servers for ARM rulesets.

## Setup

```ini
[sources.http]
type = http
url = https://packages.company.com/arm/
authToken = $HTTP_TOKEN
```

```bash
export HTTP_TOKEN="your-auth-token-here"
```

## Server Structure

```
https://packages.company.com/arm/
├── ruleset-name/
│   ├── 1.0.0/package.tar.gz
│   └── versions.json
```

## Publishing

```bash
# Create package
tar -czf package.tar.gz -C rules/ .

# Upload
curl -X PUT -T package.tar.gz \
  -H "Authorization: Bearer $HTTP_TOKEN" \
  https://packages.company.com/arm/ruleset-name/1.0.0/package.tar.gz
```

## Version Discovery

```json
{
  "versions": ["1.0.0", "1.1.0"],
  "latest": "1.1.0"
}
```

## Troubleshooting

```bash
# Test connectivity
curl -I https://packages.company.com/arm/

# Test with auth
curl -H "Authorization: Bearer $HTTP_TOKEN" \
  https://packages.company.com/arm/ruleset-name/versions.json
```

- **404**: Check URL structure and package exists
- **401**: Verify authentication token
