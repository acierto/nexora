# Design

The installer owns version compatibility and installation. --ensure requires a
stable semantic-version pin and checks only NEXORA_INSTALL_DIR, defaulting to
HOME/.local/bin. Reusing a compatible binary requires no network. Fetching the
installer itself remains a caller responsibility. Explicit installs without
--ensure can downgrade. A normal latest install remains supported.

Stage downloads inside the installation directory; verify SHA-256, require a
regular executable and compare its version against an exact requested tag before
rename. Check compatibility again before publishing an ensured install. Failures
leave an existing executable untouched and clean staging. Concurrent upgrades to
different pins are not serialized. curl/wget retain TLS checking and bounded
transfer times. No project-specific knowledge or configuration belongs here.

Integration tests build real versioned CLI binaries, package them and serve them
from local HTTP/TLS servers. Documentation shows project use and MCP path expansion.
