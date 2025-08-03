# S3 Registry Guide

## Overview

ARM's S3 registry provides reliable, globally distributed package storage using AWS S3. It leverages S3's hierarchical prefix structure for version discovery and supports custom prefixes for organization.

## Features

- **Version Discovery**: Uses S3 list-objects-v2 API to discover available versions
- **Global Distribution**: Leverages AWS S3's worldwide infrastructure
- **Prefix Support**: Organize packages with custom prefixes
- **Authentication**: AWS access keys or IAM roles
- **Scalability**: Handles large numbers of packages and versions

## Directory Structure

### Basic Structure (No Prefix)
```
s3://bucket-name/
├── packages/
│   ├── typescript-rules/
│   │   ├── 1.0.0/
│   │   │   └── typescript-rules-1.0.0.tar.gz
│   │   ├── 1.0.1/
│   │   │   └── typescript-rules-1.0.1.tar.gz
│   │   └── 1.1.0/
│   │       └── typescript-rules-1.1.0.tar.gz
│   └── security-rules/
│       ├── 2.0.0/
│       │   └── security-rules-2.0.0.tar.gz
│       └── 2.1.0/
│           └── security-rules-2.1.0.tar.gz
```

### With Organization Scope
```
s3://bucket-name/
├── packages/
│   ├── company/
│   │   ├── typescript-rules/
│   │   │   ├── 1.0.0/
│   │   │   │   └── typescript-rules-1.0.0.tar.gz
│   │   │   └── 1.0.1/
│   │   │       └── typescript-rules-1.0.1.tar.gz
│   │   └── security-rules/
│   │       └── 1.0.0/
│   │           └── security-rules-1.0.0.tar.gz
│   └── opensource/
│       └── lint-rules/
│           └── 0.1.0/
│               └── lint-rules-0.1.0.tar.gz
```

### With Custom Prefix
```
s3://bucket-name/
├── arm-registry/          # Custom prefix
│   └── packages/
│       ├── typescript-rules/
│       │   ├── 1.0.0/
│       │   │   └── typescript-rules-1.0.0.tar.gz
│       │   └── 1.0.1/
│       │       └── typescript-rules-1.0.1.tar.gz
│       └── company/
│           └── security-rules/
│               └── 1.0.0/
│                   └── security-rules-1.0.0.tar.gz
```

## Configuration Examples

### Basic S3 Registry
```ini
[sources.s3]
type = s3
bucket = my-arm-registry
region = us-east-1
authToken = ${AWS_ACCESS_KEY_ID}:${AWS_SECRET_ACCESS_KEY}
```

### S3 Registry with Prefix
```ini
[sources.s3-prod]
type = s3
bucket = company-artifacts
region = us-west-2
prefix = arm-registry
authToken = ${AWS_ACCESS_KEY_ID}:${AWS_SECRET_ACCESS_KEY}
```

## Version Discovery Process

### 1. List Request
ARM constructs an S3 list-objects-v2 request:
```
GET /?list-type=2&prefix=packages/typescript-rules/&delimiter=/
```

### 2. S3 Response
S3 returns XML with CommonPrefixes:
```xml
<ListBucketResult>
  <CommonPrefixes>
    <Prefix>packages/typescript-rules/1.0.0/</Prefix>
  </CommonPrefixes>
  <CommonPrefixes>
    <Prefix>packages/typescript-rules/1.0.1/</Prefix>
  </CommonPrefixes>
  <CommonPrefixes>
    <Prefix>packages/typescript-rules/1.1.0/</Prefix>
  </CommonPrefixes>
</ListBucketResult>
```

### 3. Version Extraction
ARM extracts versions from prefixes:
- `packages/typescript-rules/1.0.0/` → `1.0.0`
- `packages/typescript-rules/1.0.1/` → `1.0.1`
- `packages/typescript-rules/1.1.0/` → `1.1.0`

## Publishing Workflow

### Manual Upload
```bash
# Upload package to S3 with proper structure
aws s3 cp typescript-rules-1.0.0.tar.gz \
  s3://my-registry/packages/typescript-rules/1.0.0/
```

### Automated CI/CD
```yaml
# GitHub Actions example
- name: Upload to S3 Registry
  run: |
    aws s3 cp ${PACKAGE_NAME}-${VERSION}.tar.gz \
      s3://${S3_BUCKET}/packages/${PACKAGE_NAME}/${VERSION}/
```

### With Organization Scope
```bash
# Upload scoped package
aws s3 cp security-rules-1.0.0.tar.gz \
  s3://my-registry/packages/company/security-rules/1.0.0/
```

## URL Patterns

### Download URLs
- **No org**: `https://bucket.s3.region.amazonaws.com/packages/pkg/version/pkg-version.tar.gz`
- **With org**: `https://bucket.s3.region.amazonaws.com/packages/org/pkg/version/pkg-version.tar.gz`
- **With prefix**: `https://bucket.s3.region.amazonaws.com/prefix/packages/pkg/version/pkg-version.tar.gz`

### List URLs
- **No org**: `https://bucket.s3.region.amazonaws.com/?list-type=2&prefix=packages/pkg/&delimiter=/`
- **With org**: `https://bucket.s3.region.amazonaws.com/?list-type=2&prefix=packages/org/pkg/&delimiter=/`
- **With prefix**: `https://bucket.s3.region.amazonaws.com/?list-type=2&prefix=prefix/packages/pkg/&delimiter=/`

## Best Practices

### Bucket Organization
- Use dedicated buckets for ARM registries
- Apply appropriate IAM policies for access control
- Enable versioning for package history
- Configure lifecycle policies for old versions

### Prefix Strategy
- Use prefixes to separate environments (`prod/`, `staging/`)
- Organize by team or organization
- Keep prefix structure shallow for performance

### Security
- Use IAM roles instead of access keys when possible
- Apply least-privilege access policies
- Enable S3 access logging for audit trails
- Consider S3 bucket encryption

### Performance
- Use S3 Transfer Acceleration for global distribution
- Consider CloudFront for caching frequently accessed packages
- Monitor S3 request patterns and costs

## Troubleshooting

### Version Discovery Issues
1. **Check S3 permissions** - Ensure list permissions on bucket
2. **Verify prefix structure** - Confirm packages follow expected hierarchy
3. **Test list request** - Manually test S3 list-objects-v2 API
4. **Check delimiter** - Ensure proper use of `/` delimiter

### Common Problems
- **Missing versions**: Packages not uploaded to version directories
- **Permission denied**: IAM policy doesn't allow ListBucket
- **Wrong prefix**: Configuration prefix doesn't match S3 structure
- **Authentication**: Invalid or expired AWS credentials