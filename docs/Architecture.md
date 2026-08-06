# Architecture

This document describes production operation of netnotify.

## Overview

netnotify is a Go service organized around source parsers, domain notifications, a queueing application core and provider notifiers. Configuration is loaded from YAML and can be overridden by environment variables and CLI flags.

## Operational guidance

- Run as an unprivileged user.
- Store provider credentials in environment variables or a root-readable environment file.
- Keep TLS verification enabled unless using a controlled lab with private certificates.
- Monitor /health and /metrics.
- Validate configuration before restart with `netnotify validate --config /etc/netnotify/config.yaml`.
