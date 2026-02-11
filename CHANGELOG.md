# Changelog

## [Unreleased]

### Changed
- Pricing updated from $29/mo to $100/mo (still 50-70% cheaper than PlanningCenter + Gloo combined at $300-500/mo)

### Added
- Initial repository setup
- Project README and documentation
- Landing page deployed at pews.app
- Mailchimp waitlist integration (Pews Prospect tag)
- Stripe account created (acct_1SzPxQJSIrImeIRO)

### Infrastructure
- Landing page served from ctmprod via Traefik + nginx
- DNS: pews.app → 3.222.22.80, Let's Encrypt TLS
- Traefik dynamic config at /home/CTMProd/homelabCloud/traefik/config/dynamic/pews.yml
