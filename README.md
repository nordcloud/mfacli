# MFACLI - a command line tool for generating Time-based One-Time Passwords for Multi-Factor Authentication

**mfacli** allows you to generate TOTP codes from command like for different MFA clients (e.g. multiple AWS IAM Virtual MFA devices). Its most useful feature is the ability to simulate typing of the code so you can bind a command like `mfacli type CLIENT_ID` to some global shortcut and have it automatically typed into the currently focused input on a web-page or to some command-line tool which asks for the code (like awscli, serverless, etc.).

## Installation

```bash
go get github.com/nordcloud/mfacli
```

## Basic usage

### Step 1. Adding a client

```bash
mfacli add CLIENT [-s|--secret]
```

#### Set up a password

At the first execution you will be asked to set up a new password for the client secrets encrypted vault.

#### Provide a client secret

After the password is set up (or if the vault already exists) you will be asked to provide a client secret which will be used to generate TOTP codes. 

##### Secret source

if the `--secret` flag (or its short form `-s`) is omitted the value for the new secret is read from the terminal standard input without echoing the characters. If the value for the flag _is_ provided it defines the source for the new secret to be imported from. The supported forms of the flag's value are described below:

- `qr-scan`: a QR code is scanned from the screen and its decoded value is used as the new secret
- `qr-file:<IMAGE_FILE>`: a QR code is read from the `<IMAGE_FILE>` and its decoded value is used as the new secret
- `env:<ENV>`: the secret is set to the value of the `<ENV>` environment variable
- `file:<FILENAME>`: the secret is set to the whole contents of the file `<FILENAME>` (including a possible newline!)
- `pass:<PLAIN_TEXT>`: the secret is set to `<PLAIN_TEXT>`

Note: the QR code scanning from the screen assumes the `import` command from the [Imagemagick](http://imagemagick.sourceforge.net/http/www/import.html) toolkit is installed on the system.

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
