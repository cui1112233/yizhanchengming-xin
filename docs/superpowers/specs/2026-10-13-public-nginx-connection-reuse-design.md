# Public Nginx connection reuse design

## Goal

Keep the V88 repository's public Nginx configuration aligned with the active
`yzcm.site` entry point and make repeat HTTPS visits reuse their connection
state.

## Scope

- Restore the tracked HTTP-to-HTTPS `yzcm.site` servers and the `/assets/`
  forwarding rule active in production.
- Mount the existing deployment `certs` directory read-only.
- Enable a bounded TLS session cache, one-day session lifetime, TLS tickets,
  and a higher keep-alive request cap.
- Do not cache APIs, alter authentication, or change application code.

## Verification

The public-topology test will assert the certificate mount and TLS session
reuse directives. Docker Compose rendering remains valid.
