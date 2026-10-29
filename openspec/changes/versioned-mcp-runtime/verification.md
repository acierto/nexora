# Verification

Full Go tests, go vet, runtime race-detector tests, synchronized docs, strict change validation and darwin/linux arm64/amd64 release builds pass. Runtime tests exercise real HTTP release/checksum responses, immutable latest resolution per startup, offline cache pins, corruption refusal, unsupported selectors/platforms and concurrent installation. MCP integration tests run a real verified executable as a subprocess and preserve initialize responses, stdout/stderr separation and child failure codes.

The runtime loader never substitutes another version after a failed selector, download or integrity check. Configuration startup errors are owned by Nexora and do not install an agent permission gate. Missing bootstrap executables remain client launch errors. No employer-specific implementation or knowledge is included.
