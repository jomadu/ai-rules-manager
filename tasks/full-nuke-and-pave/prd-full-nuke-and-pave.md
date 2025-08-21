## Scenario: Git Registry `https://github.com/my-user/ai-rules`

This scenario demonstrates a typical development lifecycle for a ruleset repository, showing how ARM handles different types of version changes and branch workflows.

### Timeline

#### 1. Initial Release - `v1.0.0` (Tag: `1.0.0`, Commit: `1111111`)
**Branch:** `main`
**Change:** Initial release with grug-brained-dev rules

```
rules/
    cursor/
        grug-brained-dev.mdc     # ← NEW
    amazonq/
        grug-brained-dev.md      # ← NEW
```

#### 2. Patch Release - `v1.0.1` (Tag: `1.0.1`, Commit: `2222222`)
**Branch:** `main`
**Change:** Bug fix in grug-brained-dev.mdc rule

```
rules/
    cursor/
        grug-brained-dev.mdc     # ← MODIFIED (bug fix)
    amazonq/
        grug-brained-dev.md      # (unchanged)
```

#### 3. Minor Release - `v1.1.0` (Tag: `1.1.0`, Commit: `3333333`)
**Branch:** `main`
**Change:** Added task management rules (new features)

```
rules/
    cursor/
        grug-brained-dev.mdc     # (unchanged)
        generate-tasks.mdc       # ← NEW
        process-tasks.mdc        # ← NEW
    amazonq/
        grug-brained-dev.md      # (unchanged)
        generate-tasks.md        # ← NEW
        process-tasks.md         # ← NEW
```

#### 4. Pre-release - `v2.0.0-rc.1` (Tag: `2.0.0-rc.1`, Commit: `4444444`)
**Branch:** `rc`
**Change:** Breaking changes to task rules (testing phase)

```
rules/
    cursor/
        grug-brained-dev.mdc     # (unchanged)
        generate-tasks.mdc       # ← BREAKING CHANGES
        process-tasks.mdc        # ← BREAKING CHANGES
    amazonq/
        grug-brained-dev.md      # (unchanged)
        generate-tasks.md        # ← BREAKING CHANGES
        process-tasks.md         # ← BREAKING CHANGES
```

#### 5. Major Release - `v2.0.0` (Tag: `2.0.0`, Commit: `5555555`)
**Branch:** `main`
**Change:** Breaking changes merged to main (stable release)

```
rules/
    cursor/
        grug-brained-dev.mdc     # (unchanged)
        generate-tasks.mdc       # ← BREAKING CHANGES (stable)
        process-tasks.mdc        # ← BREAKING CHANGES (stable)
    amazonq/
        grug-brained-dev.md      # (unchanged)
        generate-tasks.md        # ← BREAKING CHANGES (stable)
        process-tasks.md         # ← BREAKING CHANGES (stable)
```

#### 6. Minor Release - `v2.1.0` (Tag: `2.1.0`, Commit: `6666666`)
**Branch:** `main`
**Change:** Added clean code rules (new features)

```
rules/
    cursor/
        grug-brained-dev.mdc     # (unchanged)
        generate-tasks.mdc       # (unchanged)
        process-tasks.mdc        # (unchanged)
        clean-code.mdc           # ← NEW
    amazonq/
        grug-brained-dev.md      # (unchanged)
        generate-tasks.md        # (unchanged)
        process-tasks.md         # (unchanged)
        clean-code.md            # ← NEW
```

### Git Workflow Diagram

```mermaid
gitGraph
    commit id: "1111111" tag: "v1.0.0"
    commit id: "2222222" tag: "v1.0.1"
    commit id: "3333333" tag: "v1.1.0"
    branch rc
    checkout rc
    commit id: "4444444" tag: "v2.0.0-rc.1"
    checkout main
    merge rc id: "5555555" tag: "v2.0.0"
    commit id: "6666666" tag: "v2.1.0"
```

## Setup

```sh
arm config add registry ai-rules https://github.com/my-user/ai-rules --type git

arm config add sink q --directories .amazonq/rules --rulesets ai-rules/amazonq-rules
arm config add sink cursor --directories .cursor/rules --rulesets ai-rules/cursor-rules
```

`.armrc.json`

```json
{
    "registries": {
        "ai-rules": {
            "url": "https://github.com/my-user/ai-rules",
            "type": "git"
        }
    },
    "sinks": {
        "q": {
            "directories": [
                ".amazonq/rules"
            ],
            "rulesets": [
                "ai-rules/amazonq-rules"
            ]
        },
        "cursor": {
            "directories": [
                ".cursor/rules"
            ],
            "rulesets": [
                "ai-rules/cursor-rules"
            ]
        }
    }
}
```

## Install

### Specifying No Version

When a user installs without specifying a version, ARM finds the latest semver tag, and constrains the version within the latest major.

```sh
arm install ai-rules/amazonq-rules --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules --patterns rules/cursor/*.mdc
```

`arm.json`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "version": "^2.1.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "version": "^2.1.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`arm.lock`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^2.1.0",
                "resolved": "2.1.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^2.1.0",
                "resolved": "2.1.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`./`
```
.armrc.json
arm.json
arm.lock
.cursor/
    rules/
        arm/
            ai-rules/
                cursor-rules/
                    2.1.0/
                        rules/
                            cursor/
                                grug-brained-dev.mdc
                                generate-tasks.mdc
                                process-tasks.mdc
.amazonq/
    rules/
        arm/
            ai-rules/
                amazonq-rules/
                    2.1.0/
                        rules/
                            amazonq/
                                grug-brained-dev.md
                                generate-tasks.md
                                process-tasks.md
```

`~/.arm/cache`

```
registries/
    sha256("https://github.com/my-user/ai-rules" + "git")/ <- registry cache key is hashed normalized url + normalized type
        index.json
        repository/
            ai-rules/
                .git/
                ...
        rulesets/
            sha256("rules/amazonq/*.mdc") <- ruleset cache key is hashed normalized patterns
                6666666/
                    rules/
                        amazonq/
                            grug-brained-dev.md
                            generate-tasks.md
                            process-tasks.md
            sha256("rules/cursor/*.mdc")
                6666666/
                    rules/
                        cursor/
                            grug-brained-dev.md
                            generate-tasks.md
                            process-tasks.md
```

`~/.arm/cache/registries/sha256("https://github.com/my-user/ai-rules" + "git")/index.json`

```json
{
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "last_accessed_on": "2024-01-15T10:30:00Z",
  "normalized_registry_url": "https://github.com/user/repo",
  "normalized_registry_type": "git",
  "rulesets": {
    "xyz789abc123...": {
      "normalized_ruleset_patterns": ["rules/cursor/*.mdc"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "6666666": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    },
    "aadbbf3222b3...": {
      "normalized_ruleset_patterns": ["rules/amazonq/*.mdc"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "6666666": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    }
  }
}
```

### Specifying Full Version

specifying the full version string will pin (e.g., `=1.0.0`)

```sh
arm install ai-rules/amazonq-rules@1.0.0 --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules@1.0.0 --patterns rules/cursor/*.mdc
```

`arm.json`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "version": "=1.0.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "version": "=1.0.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`arm.lock`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "=1.0.0",
                "resolved": "1.0.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "=1.0.0",
                "resolved": "1.0.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`./`
```
.armrc.json
arm.json
arm.lock
.cursor/
    rules/
        arm/
            ai-rules/
                cursor-rules/
                    1.0.0/
                        rules/
                            cursor/
                                grug-brained-dev.mdc
.amazonq/
    rules/
        arm/
            ai-rules/
                amazonq-rules/
                    1.0.0/
                        rules/
                            amazonq/
                                grug-brained-dev.md
```

`~/.arm/cache`

```
registries/
    sha256("https://github.com/my-user/ai-rules" + "git")/
        index.json
        repository/
            ai-rules/
                .git/
                ...
        rulesets/
            sha256("rules/amazonq/*.md")
                1111111/
                    rules/
                        amazonq/
                            grug-brained-dev.md
            sha256("rules/cursor/*.mdc")
                1111111/
                    rules/
                        cursor/
                            grug-brained-dev.mdc
```

`~/.arm/cache/registries/sha256("https://github.com/my-user/ai-rules" + "git")/index.json`

```json
{
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "last_accessed_on": "2024-01-15T10:30:00Z",
  "normalized_registry_url": "https://github.com/user/repo",
  "normalized_registry_type": "git",
  "rulesets": {
    "xyz789abc123...": {
      "normalized_ruleset_patterns": ["rules/cursor/*.mdc"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "1111111": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    },
    "aadbbf3222b3...": {
      "normalized_ruleset_patterns": ["rules/amazonq/*.md"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "1111111": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    }
  }
}
```

### Specifying Major and Minor Version

specifying the major and minor will pin within minor (e.g., `~1.0.0`)

```sh
arm install ai-rules/amazonq-rules@1.0 --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules@1.0 --patterns rules/cursor/*.mdc
```

`arm.json`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "version": "~1.0.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "version": "~1.0.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`arm.lock`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "~1.0.0",
                "resolved": "1.0.1",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "~1.0.0",
                "resolved": "1.0.1",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`./`
```
.armrc.json
arm.json
arm.lock
.cursor/
    rules/
        arm/
            ai-rules/
                cursor-rules/
                    1.0.1/
                        rules/
                            cursor/
                                grug-brained-dev.mdc
.amazonq/
    rules/
        arm/
            ai-rules/
                amazonq-rules/
                    1.0.1/
                        rules/
                            amazonq/
                                grug-brained-dev.md
```

`~/.arm/cache`

```
registries/
    sha256("https://github.com/my-user/ai-rules" + "git")/
        index.json
        repository/
            ai-rules/
                .git/
                ...
        rulesets/
            sha256("rules/amazonq/*.md")
                2222222/
                    rules/
                        amazonq/
                            grug-brained-dev.md
            sha256("rules/cursor/*.mdc")
                2222222/
                    rules/
                        cursor/
                            grug-brained-dev.mdc
```

`~/.arm/cache/registries/sha256("https://github.com/my-user/ai-rules" + "git")/index.json`

```json
{
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "last_accessed_on": "2024-01-15T10:30:00Z",
  "normalized_registry_url": "https://github.com/user/repo",
  "normalized_registry_type": "git",
  "rulesets": {
    "xyz789abc123...": {
      "normalized_ruleset_patterns": ["rules/cursor/*.mdc"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "2222222": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    },
    "aadbbf3222b3...": {
      "normalized_ruleset_patterns": ["rules/amazonq/*.md"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "2222222": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    }
  }
}
```

### Specifying Major

specifying the major will allow within major (e.g., `^1.0.0`)

```sh
arm install ai-rules/amazonq-rules@1 --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules@1 --patterns rules/cursor/*.mdc
```

`arm.json`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "version": "^1.0.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "version": "^1.0.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`arm.lock`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^1.0.0",
                "resolved": "1.1.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^1.0.0",
                "resolved": "1.1.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`./`
```
.armrc.json
arm.json
arm.lock
.cursor/
    rules/
        arm/
            ai-rules/
                cursor-rules/
                    1.1.0/
                        rules/
                            cursor/
                                grug-brained-dev.mdc
                                generate-tasks.mdc
                                process-tasks.mdc
.amazonq/
    rules/
        arm/
            ai-rules/
                amazonq-rules/
                    1.1.0/
                        rules/
                            amazonq/
                                grug-brained-dev.md
                                generate-tasks.md
                                process-tasks.md
```

`~/.arm/cache`

```
registries/
    sha256("https://github.com/my-user/ai-rules" + "git")/
        index.json
        repository/
            ai-rules/
                .git/
                ...
        rulesets/
            sha256("rules/amazonq/*.md")
                3333333/
                    rules/
                        amazonq/
                            grug-brained-dev.md
                            generate-tasks.md
                            process-tasks.md
            sha256("rules/cursor/*.mdc")
                3333333/
                    rules/
                        cursor/
                            grug-brained-dev.mdc
                            generate-tasks.mdc
                            process-tasks.mdc
```

`~/.arm/cache/registries/sha256("https://github.com/my-user/ai-rules" + "git")/index.json`

```json
{
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "last_accessed_on": "2024-01-15T10:30:00Z",
  "normalized_registry_url": "https://github.com/user/repo",
  "normalized_registry_type": "git",
  "rulesets": {
    "xyz789abc123...": {
      "normalized_ruleset_patterns": ["rules/cursor/*.mdc"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "3333333": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    },
    "aadbbf3222b3...": {
      "normalized_ruleset_patterns": ["rules/amazonq/*.md"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "3333333": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    }
  }
}
```

### Specifying Branch

specifying a branch name will track the head of the branch

```sh
arm install ai-rules/amazonq-rules@main --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules@main --patterns rules/cursor/*.mdc
```

`arm.json`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "version": "main",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "version": "main",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`arm.lock`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "main",
                "resolved": "6666666",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "main",
                "resolved": "6666666",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`./`
```
.armrc.json
arm.json
arm.lock
.cursor/
    rules/
        arm/
            ai-rules/
                cursor-rules/
                    6666666/
                        rules/
                            cursor/
                                grug-brained-dev.mdc
                                generate-tasks.mdc
                                process-tasks.mdc
                                clean-code.mdc
.amazonq/
    rules/
        arm/
            ai-rules/
                amazonq-rules/
                    6666666/
                        rules/
                            amazonq/
                                grug-brained-dev.md
                                generate-tasks.md
                                process-tasks.md
                                clean-code.md
```

`~/.arm/cache`

```
registries/
    sha256("https://github.com/my-user/ai-rules" + "git")/
        index.json
        repository/
            ai-rules/
                .git/
                ...
        rulesets/
            sha256("rules/amazonq/*.md")
                6666666/
                    rules/
                        amazonq/
                            grug-brained-dev.md
                            generate-tasks.md
                            process-tasks.md
                            clean-code.md
            sha256("rules/cursor/*.mdc")
                6666666/
                    rules/
                        cursor/
                            grug-brained-dev.mdc
                            generate-tasks.mdc
                            process-tasks.mdc
                            clean-code.mdc
```

`~/.arm/cache/registries/sha256("https://github.com/my-user/ai-rules" + "git")/index.json`

```json
{
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "last_accessed_on": "2024-01-15T10:30:00Z",
  "normalized_registry_url": "https://github.com/user/repo",
  "normalized_registry_type": "git",
  "rulesets": {
    "xyz789abc123...": {
      "normalized_ruleset_patterns": ["rules/cursor/*.mdc"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "6666666": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    },
    "aadbbf3222b3...": {
      "normalized_ruleset_patterns": ["rules/amazonq/*.md"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "6666666": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    }
  }
}
```

## Uninstall

```sh
arm install ai-rules/amazonq-rules --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules --patterns rules/cursor/*.mdc
arm uninstall ai-rules/cursor-rules
```

`arm.json`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "version": "^2.1.0",
                "patterns": ["rules/amazonq/*.md"]
            }
        }
    }
}
```

`arm.lock`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^2.1.0",
                "resolved": "2.1.0",
                "patterns": ["rules/amazonq/*.md"]
            }
        }
    }
}
```

`./`
```
.armrc.json
arm.json
arm.lock
.cursor/
    rules/
        (arm directory removed - no rulesets installed)
.amazonq/
    rules/
        arm/
            ai-rules/
                amazonq-rules/
                    2.1.0/
                        rules/
                            amazonq/
                                grug-brained-dev.md
                                generate-tasks.md
                                process-tasks.md
                                clean-code.md
```

`~/.arm/cache`

```
registries/
    sha256("https://github.com/my-user/ai-rules" + "git")/
        index.json
        repository/
            ai-rules/
                .git/
                ...
        rulesets/
            sha256("rules/amazonq/*.md")
                6666666/
                    rules/
                        amazonq/
                            grug-brained-dev.md
                            generate-tasks.md
                            process-tasks.md
                            clean-code.md
            sha256("rules/cursor/*.mdc") <- cache maintained
                6666666/
                    rules/
                        cursor/
                            grug-brained-dev.mdc
                            generate-tasks.mdc
                            process-tasks.mdc
                            clean-code.mdc
```

`~/.arm/cache/registries/sha256("https://github.com/my-user/ai-rules" + "git")/index.json`

```json
{
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "last_accessed_on": "2024-01-15T10:30:00Z",
  "normalized_registry_url": "https://github.com/user/repo",
  "normalized_registry_type": "git",
  "rulesets": {
    "xyz789abc123...": {
      "normalized_ruleset_patterns": ["rules/cursor/*.mdc"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "6666666": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    },
    "aadbbf3222b3...": {
      "normalized_ruleset_patterns": ["rules/amazonq/*.md"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "6666666": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    }
  }
}
```

## Outdated

```sh
arm install ai-rules/amazonq-rules@1 --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules@1 --patterns rules/cursor/*.mdc
arm outdated

| Registry | Ruleset       | Current | Wanted | Latest  |
|----------|---------------|---------|--------|---------|
| ai-rules | amazonq-rules | 1.1.0   | 1.1.0  | 2.1.0   |
| ai-rules | cursor-rules  | 1.1.0   | 1.1.0  | 2.1.0   |
```

## Update

### 1. After Release of 1.0.0, Prior to Release of 1.1.0

```sh
arm install ai-rules/amazonq-rules@1 --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules@1 --patterns rules/cursor/*.mdc
arm outdated

All rulesets are up to date!
```

`arm.json`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "version": "^1.0.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "version": "^1.0.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`arm.lock`

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^1.0.0",
                "resolved": "1.0.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^1.0.0",
                "resolved": "1.0.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`./`
```
.armrc.json
arm.json
arm.lock
.cursor/
    rules/
        arm/
            ai-rules/
                cursor-rules/
                    1.0.0/
                        rules/
                            cursor/
                                grug-brained-dev.mdc
.amazonq/
    rules/
        arm/
            ai-rules/
                amazonq-rules/
                    1.0.0/
                        rules/
                            amazonq/
                                grug-brained-dev.md
```

### 2. After Release of 1.1.0

```sh
arm outdated

| Registry | Ruleset       | Current | Wanted | Latest  |
|----------|---------------|---------|--------|---------|
| ai-rules | amazonq-rules | 1.0.0   | 1.1.0  | 1.1.0   |
| ai-rules | cursor-rules  | 1.0.0   | 1.1.0  | 1.1.0   |

arm update
```

`arm.json` (unchanged)

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "version": "^1.0.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "version": "^1.0.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`arm.lock` (updated resolved versions)

```json
{
    "rulesets": {
        "ai-rules": {
            "amazonq-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^1.0.0",
                "resolved": "1.1.0",
                "patterns": ["rules/amazonq/*.md"]
            },
            "cursor-rules": {
                "url": "https://github.com/my-user/ai-rules",
                "type": "git",
                "constraint": "^1.0.0",
                "resolved": "1.1.0",
                "patterns": ["rules/cursor/*.mdc"]
            },
        }
    }
}
```

`./` (updated with new version directories)
```
.armrc.json
arm.json
arm.lock
.cursor/
    rules/
        arm/
            ai-rules/
                cursor-rules/
                    1.1.0/
                        rules/
                            cursor/
                                grug-brained-dev.mdc
                                generate-tasks.mdc
                                process-tasks.mdc
.amazonq/
    rules/
        arm/
            ai-rules/
                amazonq-rules/
                    1.1.0/
                        rules/
                            amazonq/
                                grug-brained-dev.md
                                generate-tasks.md
                                process-tasks.md
```

`~/.arm/cache` (new version cached)

```
registries/
    sha256("https://github.com/my-user/ai-rules" + "git")/
        index.json
        repository/
            ai-rules/
                .git/
                ...
        rulesets/
            sha256("rules/amazonq/*.md")
                1111111/
                    rules/
                        amazonq/
                            grug-brained-dev.md
                3333333/
                    rules/
                        amazonq/
                            grug-brained-dev.md
                            generate-tasks.md
                            process-tasks.md
            sha256("rules/cursor/*.mdc")
                1111111/
                    rules/
                        cursor/
                            grug-brained-dev.mdc
                3333333/
                    rules/
                        cursor/
                            grug-brained-dev.mdc
                            generate-tasks.mdc
                            process-tasks.mdc
```

## Info

```sh
arm install ai-rules/amazonq-rules@1 --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules@1 --patterns rules/cursor/*.mdc
arm info ai-rules/amazonq-rules
```

```
Ruleset: ai-rules/amazonq-rules
Registry: https://github.com/my-user/ai-rules (git)
Patterns:
  - rules/amazonq/*.md
Installed:
  - .amazonq/rules/arm/ai-rules/amazonq-rules/1.1.0/
Sinks:
  - q
Constraint: ^1.0.0
Resolved: 1.1.0
Wanted: 1.1.0
Latest: 2.1.0
```

```sh
arm info
```

```
ai-rules/amazonq-rules
  Registry: https://github.com/my-user/ai-rules (git)
  Patterns:
    - rules/amazonq/*.md
  Installed:
    - .amazonq/rules/arm/ai-rules/amazonq-rules/1.1.0/
  Sinks:
    - q
  Constraint: ^1.0.0 | Resolved: 1.1.0 | Wanted: 1.1.0 | Latest: 2.1.0

ai-rules/cursor-rules
  Registry: https://github.com/my-user/ai-rules (git)
  Patterns:
    - rules/cursor/*.mdc
  Installed:
    - .cursor/rules/arm/ai-rules/cursor-rules/1.1.0/
  Sinks:
    - cursor
  Constraint: ^1.0.0 | Resolved: 1.1.0 | Wanted: 1.1.0 | Latest: 2.1.0
```

## Version

```sh
arm version
```

```
arm 1.2.3
commit: a1b2c3d4
arch: darwin/arm64
```

## Help

```sh
arm help
```

```
Usage: arm <command> [options]

Commands:
  install <ruleset>     Install a ruleset
  uninstall <ruleset>   Uninstall a ruleset
  update [ruleset]      Update rulesets
  outdated              Show outdated rulesets
  list                  List installed rulesets
  info [ruleset]        Show ruleset information
  config                Manage configuration
  version               Show version
  help                  Show help

Options:
  --help               Show command help
```

## List

```sh
arm install ai-rules/amazonq-rules@1 --patterns rules/amazonq/*.md
arm install ai-rules/cursor-rules@1 --patterns rules/cursor/*.mdc
arm list
```

```
ai-rules/amazonq-rules@1.1.0
ai-rules/cursor-rules@1.1.0
```
