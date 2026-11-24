# Koschei ARVIS Open-Source Components

This directory contains reusable Solana security infrastructure that can be inspected and used independently from the hosted Koschei dashboard.

## Packages

### `event-normalizer`

A dependency-free TypeScript package that normalizes Pump-style and Raydium-style observations into a deterministic, versioned event contract.

```bash
cd oss/event-normalizer
npm ci
npm run check
npm test
npm pack --dry-run
```

### `schemas`

Machine-readable contracts for ARVIS developer integrations.

Current schema:

- `signed-verdict.schema.json`

## Design rules

1. Normalization and risk scoring remain separate layers.
2. Missing timestamps, targets or source classification are rejected.
3. A final verdict is trusted only when its evidence and rule metadata are present.
4. AI-generated explanations never replace deterministic evidence.
5. Open-source components must not expose production secrets or private infrastructure credentials.

## License

MIT. See the repository root [LICENSE](../LICENSE) file.
