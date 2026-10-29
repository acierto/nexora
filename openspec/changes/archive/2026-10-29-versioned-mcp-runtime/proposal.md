# Why

Projects currently launch whichever Nexora executable is on PATH and duplicate startup warnings. They need a shared runtime selector with exact pins and an explicit latest option.

# What Changes

Add mcp --version latest|vX.Y.Z. Nexora resolves latest to an immutable release, verifies and caches official binaries, then serves through the selected version. Startup diagnostics belong upstream; repository tools remain usable when MCP setup fails. Installation remains a separate prerequisite.

# Capabilities

## Modified Capabilities

- mcp-server: deterministic runtime selection and actionable startup diagnostics.

# Impact

Internal runtime selection package, CLI, tests, agent/install/reference docs and release archives. No dependencies, remote serving, automatic install of the launcher or changes to knowledge authorization.
