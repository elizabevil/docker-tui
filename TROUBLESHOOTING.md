# Troubleshooting

## Git push to GitHub fails: `GnuTLS, handshake failed: The TLS connection was non-properly terminated.`

### Symptom

```
$ git push origin <branch>
fatal: unable to access 'https://github.com/<user>/<repo>.git/':
GnuTLS, handshake failed: The TLS connection was non-properly terminated.
```

### Environment

- Debian / `git version 2.47.3` built with **only the `gnutls` SSL backend**
  (`git config http.sslBackend openssl` → `fatal: Unsupported SSL backend 'openssl'. Supported SSL backends: gnutls`)
- HTTPS via `curl` (OpenSSL) to the same host works
- TCP `nc -zv github.com 443` → `open` (TCP handshake succeeds)

### Root cause

The TCP connection to `github.com:443` succeeds, but the **GnuTLS TLS handshake**
fails to complete (ClientHello sent, no ServerHello received). On this network,
GnuTLS specifically cannot negotiate TLS with GitHub while OpenSSL-based clients
(`curl`) succeed on the same path.

Likely culprits, in order of probability:

1. **MTU / Path MTU Discovery issue** — GnuTLS may be sending a ClientHello that
   exceeds the path MTU and is being silently dropped (no ICMP back).
2. **ISP / transparent proxy / DPI** interfering with TLS ClientHello bytes.
3. **Outdated `libgnutls30`** on Debian.
4. **Corporate firewall** rewriting TLS.

The fact that `curl` (OpenSSL) works on the exact same path strongly points to
a GnuTLS-specific client behavior, not a network outage.

### Diagnostic checklist

```bash
# 1. TCP reaches the host
nc -zv github.com 443

# 2. OpenSSL-based TLS works
curl -vI https://github.com

# 3. Which SSL backends does this git support?
git config http.sslBackend openssl   # → "Unsupported" on gnutls-only builds

# 4. GnuTLS-only git reproduces the hang
GIT_CURL_VERBOSE=1 git push origin <branch>
```

### Fix (applied)

Switch the remote from HTTPS to SSH. SSH uses its own protocol stack (OpenSSH)
and bypasses the broken GnuTLS path.

```bash
git remote set-url origin git@github.com:<user>/<repo>.git

# One-time: import GitHub host keys (StrictHostKeyChecking requires this)
ssh-keyscan -T 15 -t rsa,ed25519,ecdsa github.com \
  | grep -v '^#' >> ~/.ssh/known_hosts

git push origin <branch>
```

Confirmed: push of 144 objects / 108 written completed via SSH; force-update
`d63bcac → 3cb9198` on `master-builder`.

### Other workarounds (not used)

- Lower interface MTU and retry: `sudo ip link set dev wlp3s0 mtu 1400`
- Upgrade: `sudo apt install --only-upgrade git libgnutls30`
- Install a git build with OpenSSL backend (e.g. from
  `git-core` Debian backports or build from source with
  `USE_OPENSSL=1`).
- Last resort: `git config --global http.sslVerify false` (insecure — does
  not fix handshake, only skips cert verification).

### Prevention

- Prefer SSH remotes on networks where GnuTLS handshakes are unreliable.
- If SSH is unavailable, ensure `git` is built with the OpenSSL backend.