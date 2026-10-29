# Regular expression policy filters

Tag retention, tag immutability, replication, and P2P preheat support an explicit
pattern engine for each repository, tag, or label pattern. Select **RegEx** in the
portal or set `kind` to `regex` in the API. An omitted kind selects the existing
Doublestar engine. Existing policies require no migration or database changes.

## Matching and examples

Regex filters use Go's `regexp` syntax (RE2). Each expression matches the entire
value, including when it contains inline flags or alternation. An empty pattern
matches everything. Expressions are case sensitive unless `(?i)` is specified,
limited to 512 Unicode characters, and compiled when their selectors are loaded.
Invalid expressions are rejected when creating or updating a policy.

| Intent | Pattern | Decoration |
| --- | --- | --- |
| Final release tags | `v?\d+\.\d+\.\d+` | matching |
| CI tags from git describe | `v\d+\.\d+\.\d+-\d+-g[0-9a-f]+` | matching |
| Skip source variants | `.*-source` | excluding |
| Latest, regardless of case | `(?i)latest` | matching |
| Repositories under library | `library/.*` | matching |

For example, the final release expression matches `v1.9.0` and `1.10.3`, and skips
`v2.0.0-rc.1` and `v1.9.0-12-ge7e744a7`. Use `.*` explicitly when a substring should
match. Named capture groups are accepted but their captured values have no policy
semantics. Lookahead, lookbehind, and backreferences are unsupported; use an
excluding filter to omit matching values.

Untagged artifacts are selected using the existing `untagged` flag, independently
of whether the expression could match an empty string. Regex quantifiers, commas,
and braces are saved verbatim. Doublestar retains its existing comma-list input.
Use retention's dry-run before enabling a new deletion policy.

## API examples

Retention and immutability reuse their existing selector structure:

```json
{"kind":"regex","decoration":"matches","pattern":"v?\\d+\\.\\d+\\.\\d+","extras":"{\"untagged\":false}"}
```

Replication adds the optional `kind` field to each pattern filter:

```json
{"type":"tag","kind":"regex","decoration":"matches","value":"v?\\d+\\.\\d+\\.\\d+"}
```

Preheat includes the engine in the existing filter JSON:

```json
[{"type":"repository","kind":"doublestar","value":"library/**"},{"type":"tag","kind":"regex","value":"v?\\d+\\.\\d+\\.\\d+"}]
```

Legacy label filters retain exact label-list matching. A regex label filter
matches label names; multiple supplied replication label expressions each need a
matching label. In the portal a regex label expression is entered as raw text.

A regex repository filter can require a full remote catalog listing, which may
be slower than an exact name or a Doublestar prefix. Registry adapters that cannot
list repositories, such as GitHub Container Registry, still require a specific
repository; literal regex expressions can use that exact-repository path.

## Release and downgrade notes

Before rolling back to a core without regex support, remove regex policies or
replace them with reviewed Doublestar equivalents. Older retention implementations
reject the unknown selector and stop the job. Older immutability implementations
can swallow selector errors and silently allow an otherwise protected tag to be
overwritten. Older replication implementations ignore the additional `kind`
field and interpret its pattern as Doublestar, which can change the replicated
set. Preheat also needs its regex filters converted before rollback.

The new implementation propagates immutability evaluation errors to prevent an
overwrite or deletion when a rule cannot be evaluated.
