# Notable rationale of deployer toolkit

## command line flags convention

The canonical representation of flags in this package is:
* single-dash for one-char flags (-v, -h)
* double-dash for multi-char flags (--foo, --long-option)
pflag allows one-char to have one or two dashes.

The tooling we configure all use `pflag`, so our canonical
representation is more restrictive of those.
