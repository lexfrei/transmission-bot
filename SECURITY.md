# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

Please report security vulnerabilities to: f@lex.la

**Do not** report security vulnerabilities through public GitHub issues.

## Security Features

- **Whitelist-based access control**: Only configured Telegram user IDs can interact with the bot
- **Minimal container image**: Built from scratch with only necessary components
- **Non-root execution**: Container runs as unprivileged user (nobody:65534)
- **No secrets in logs**: Sensitive data is not logged

## Best Practices

When deploying this bot:

1. **Secure Transmission RPC**
   - Enable authentication on Transmission RPC
   - Use HTTPS if possible
   - Restrict network access to Transmission

2. **Protect Bot Token**
   - Store the Telegram bot token securely
   - Use environment variables or secrets management
   - Never commit tokens to version control

3. **Limit Access**
   - Only add trusted Telegram user IDs to the whitelist
   - Regularly review the allowed users list

4. **Keep Updated**
   - Update to the latest version regularly
   - Monitor for security advisories
