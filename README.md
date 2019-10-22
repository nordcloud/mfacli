# MFACLI - a command line tool for generating Time-based One-Time Passwords for Multi-Factor Authentication

**mfacli** allows you to generate TOTP codes from command like for different MFA clients (e.g. multiple AWS IAM Virtual MFA devices). Its most useful feature is the ability to simulate typing of the code so you can bind a command like `mfacli type CLIENT_ID` to some global shortcut and have it automatically typed into the currently focused input on a web-page or to some command-line tool which asks for the code (like awscli).

## Installation

```bash
go get github.com/nordcloud/mfacli
```

## Basic usage

### Step 1. Adding a client

```bash
mfacli add CLIENT [--secret]
```

#### Set up a password

At the first execution you will be asked to set up a new password for the client secrets encrypted vault.

#### Provide a client secret

After the password is set up (or if the vault already exists) you will be asked to provide a client secret which will be used to generate TOTP codes. The characters won't be echoed on the terminal (like in sudo).

### Step 2. Generate the TOTP code

#### Print to standard output

```bash
mfacli print CLIENT_ID [--newline]
```

#### Copy to clipboard (currently only supported on Linux with X.org)

```bash
mfacli clipboard CLIENT_ID [--newline]
```

#### Simulate typing (currently only supported on Linux with X.org)

```bash
mfacli type CLIENT_ID [--newline]
```

## How it works

All client secrets are stored in an encrypted file which is called a vault. Its default location is `~/.mfacli/mfacli.vault` though a custom value can be provided using `--vault` flag (see `mfacli --help` for details).

To prevent typing the vault password every time you want to generate a TOTP code only the first execution of **mfacli** asks for password. It then starts a secrets cache server (using the encryption key which is SHA-256 sum of the password) which listens on a Unix socket (`~/.mfacli/mfacli.sock` by default). Upon all subsequent executions **mfacli** connects to the socket to retrieve the secret and then generates the code based on it. This way the secrets are never stored on disk unencrypted.


