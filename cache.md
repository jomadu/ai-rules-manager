# cache directory structure

cache/
    registries/
        hash(registry_url + registry_type)/ < -- git based registry
            index.json
            repository/ < -- cloned git repository, if authToken, apiType, and apiVersion not specified
            rulesets/
                hash(normalized_ruleset_patterns)/
                    path/to/rules/
                        ...
        hash(registry_url + registry_type)/
            index.json
            rulesets/
                hash(ruleset_name)/
                    ...
