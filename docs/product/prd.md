version spec is source/ruleset@version

Configuration: .armrc (user/project config)
Manifest: arm.json (dependencies)
Lock file: arm.lock (resolved versions)

```sh
// Install rulesets
arm install source/ruleset@version, ...
arm install

// Remove a ruleset
arm uninstall source/ruleset@version, ...
arm uninstall

// Update rulesets
arm update source/ruleset@version, ...
arm update

// List installed rulesets
arm list [--format=table\|json]

// Show outdated rulesets
arm outdated

// Manage configuration
arm config [list\|get\|set] [key] [value]

// Clean cache and unused files
arm clean

// Show help
arm help

// Show version
arm version
```

`~/.arm/.armrc`
```
[registries]
// default used when a package version spec doesnt include the source. must be configured by user, since there is no authoritative central rules registry (like dockerhub.com)
default = git://www.github.com/dk/bongo-registry

[git]
// defaults for the git based registries (both api driven and git operations driven)
concurrency=1
rateLimit=10

[s3]
// defaults for s3 based registries
concurrency=10
rateLimit=100

[gitlab]
// defaults for gitlab based registries
concurrency=2
rateLimit=60
```

`~/.arm/arm.json`
```
{
    "channels": {
        "q": {
            "directories": ["~/.aws/amazonq/rules"]
        }
    },
    "rulesets": {
        "bongo-rules": {
            "version": "^1.0.0", // installs the bongo-rules ruleset from the default dk registry following semver range, since source wasn't specified
            "channels": [
                "q"
            ]
    }
}
```

```
~/aws
    amazonq
        rules
            arm
                default
                    bongo-rules
                        1.0.1
                            tappin.mdc
                            slappin.mdc
```

`./.armrc`
```ini
[registries]
// default used when a package version spec doesnt include the source. must be configured by user, since there is no authoritative central rules registry (like dockerhub.com). overrides ~/.arm/.armrc default
default = git://www.github.com/mario/here-we-go-rules-registry
// named registries
awesome-cursorrules = git://www.github.com/PatrickJS/awesome-cursorrules
cursor-rules = git://www.github.com/sparsesparrow/cursor-rules
peach = s3://peach.us-east-1.amazonaws.com/
toad = s3://toad.us-east-1.amazonaws.com/
kart = gitlab://gitlab.yoshi.com/project/1234
mario = gitlab://gitlab.wario.com/group/5678

[registries.awesome-cursorrules]
// for git type registries, if authToken, apiType, or apiVersion are not provided, uses git operations with users git auth
// concurrency and rate limit on specific registries overrides the default configured for the registry type
concurrency = 2
rateLimit = 10

[registries.cursor-rules]
authToken = $GITHUB_PAT // for git type registry if authToken, apiType and apiVersion are specified, uses the specific api to retrieve files (not all will be supported, but some will)
apiType = github
apiVersion = 2022-11-28

[registries.peach]
// omission of profile uses default aws profile
prefix = /registries/panda-bear

[registries.toad]
profile = toad // uses a toad aws profile
// ommision of prefix uses no prefix

[registries.kart]
authToken = $GITLAB_KART_RULES_TOKEN
apiVersion = 3

[registries.mario]
authToken = $GITLAB_MARIO_TOKEN
// omission of apiVersion defaults to the latest version of the api (4 for gitlab at the time of writing)

[gitlab]
// defaults for gitlab based registries in this directory, overrides the ~/.arm/.armrc config
concurrency=2
rateLimit=60
```

`./arm.json`
```
{
    "channels": {
        "cursor": {
            "directories": [".cursor/rules"]
        }
        "q": {
            "directories": [".amazonq/rules"]
        }
    },
    "rulesets": {
        "wahoo-rules": {
            "version": "^1.0.0", // installs the wahoo-rules ruleset from the default registry following semver range, since source wasn't specified
            "channels": [
                "q"
            ]
        }
        "awesome-cursorrules/rules-new-python": {
            "version": "latest", // tracks the latest changes to the default branch of the project. rules-new/python-*.mdc is the glob pattern selecting the set of files to install
            // for git based registries, the scope of the package is defined by a set of glob patterns
            "matchingPatterns": [
                "rules-new/python-*.mdc"
            ]
            "channels": [
                "cursor"
            ]
        }
        "cursor-rules/base-devops": {
            "version": "main" // tracks latest changes to a named branch
            "matchingPatterns": [
                ".cursor/rules/01-base-devops.mdc"
            ]
            "channels": [
                "cursor"
            ]
        }
        "cursor-rules/base-agentic": {
            "version": "^1.0.0", // tracks semver tags, supports tagging with v1.0.0 and 1.0.0 syntax
            "matchingPatterns": [
                ".cursor/rules/01-base-agentic.mdc"
            ]
            "channels": [
                "cursor"
            ]
        }
        "cursor-rules/inspirations": {
            "version": "53c5307", // pinned to a commit
            "matchingPatterns": {
                ".cursor/rules/inspirations.mdc"
            }
            "channels": [
                "cursor"
            ]
        }
        "peach/dress-rules": {
            "version": "~1.0.0", // supports ^, ~, >=, <=, >, <, =
            "channels": [
                "q"
            ]
        }
        "toad/shroom-rules": {
            "version": ">=2.0.0",
            "channels": [
                "q"
            ]
        }
        "kart/mechanic-rules": {
            "version": "~3.0.0",
            "channels": [
                "q"
            ]
        }
        "mario/jumping-rules": {
            "version": "^1.1.0"
            "channels": [
                "q"
            ]
        }
    }
}
```

```
./
    .amazonq
        rules
            default
                wahoo-rules
                    1.1.0
                        wa.mdc
                        hoo.mdc
            peach
                dress-rules
                    1.0.4
                        accessories.md
                        colors.md
            toad
                shroom-rules
                    1.6.1
                        bad-shrooms.md
                        good-shrooms.md
            kart
                mechanic-rules
                    3.0.2
                        axle-work.md
                        tires.md
            mario
                jumping-rules
                    1.7.0
                        form.md
                        knee-drive.md
    .cursor
        rules
            awesome-cursorrules
                rules-new-python
                    latest
                        rules-new
                            python-sec.mdc
                            python-dev.mdc
            cursor-rules
                base-devops
                    main
                        .cursor
                            rules
                                01-base-devops.mdc
                base-agentic
                    1.2.0
                        .cursor
                            rules
                                01-base-devops.mdc
                inspirations
                    53c5307
                        .cursor
                            rules
                                inspirations.mdc
```


```
~/.arm/cache
    registries
        {registry}
            repository
                ...
            {package}
                {version}
                    ruleset.tar.gz
            metadata.json
            versions.json
```

s3 structure
```
prefix/
    package/
        version/
            ruleset.tar.gz
```
