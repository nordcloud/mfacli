package secret

import (
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

type SecretValue struct {
	value string
	isSet bool
}

func (s *SecretValue) String() string {
	return s.value
}

func (s *SecretValue) Set(arg string) error {
	var err error

	if val, ok := stripPrefix(arg, "pass:"); ok {
		s.value = val
	} else if val, ok := stripPrefix(arg, "file:"); ok {
		err = s.setFromFile(val)
	} else if val, ok := stripPrefix(arg, "env:"); ok {
		err = s.setFromEnv(val)
	} else if arg == "qr-scan" {
		err = s.setFromQRScan()
	} else if val, ok := stripPrefix(arg, "qr-file:"); ok {
		err = s.setFromQRFile(val)
	} else {
		err = errors.Errorf("Invalid secret format")
	}

	if err == nil {
		s.isSet = true
	}

	return err
}

func (s *SecretValue) Type() string {
	return "secret"
}

func (s *SecretValue) ReadSecret(prompt, confirmPrompt string) (string, error) {
	if s.isSet {
		return s.value, nil
	}

	value, err := readSecret(prompt)
	if confirmPrompt == "" || err != nil {
		return value, err
	}

	var confirmation string
	for value != confirmation {
		confirmation, err = readSecret(confirmPrompt)
		if err != nil {
			return "", err
		}
	}
	return value, nil
}

func (s *SecretValue) IsSet() bool {
	return s.isSet
}

func (s *SecretValue) setFromFile(filename string) error {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	s.value = string(body)
	return nil
}

func (s *SecretValue) setFromEnv(env string) error {
	val, ok := os.LookupEnv(env)
	if !ok {
		return errors.Errorf("Secret env %s is not set", env)
	}
	s.value = val
	return nil
}

func (s *SecretValue) setFromQRScan() error {
	tmpFile, err := ioutil.TempFile("", "mfacli-img*.png")
	if err != nil {
		return err
	}
	filename := tmpFile.Name()
	defer os.Remove(filename)

	if err := exec.Command("import", filename).Run(); err != nil {
		return err
	}

	return s.setFromQRFile(filename)
}

func (s *SecretValue) setFromQRFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return err
	}

	qrReader := qrcode.NewQRCodeReader()
	res, err := qrReader.DecodeWithoutHints(bmp)
	if err != nil {
		return errors.Wrap(err, "failed to decode qr code")
	}

	s.value = tryParseUrl(res.String())
	return nil
}

func tryParseUrl(raw string) string {
	url, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if url.Scheme != "otpauth" {
		return raw
	}

	secret := url.Query().Get("secret")
	if secret == "" {
		return raw
	}

	return secret
}

func readSecret(prompt string) (string, error) {
	for {
		fmt.Fprint(os.Stderr, prompt)
		data, err := terminal.ReadPassword(0)
		fmt.Fprint(os.Stderr, "\n")
		if err != nil {
			return "", err
		}

		if len(data) > 0 {
			return string(data), nil
		}

		fmt.Fprintln(os.Stderr, "Empty input is not allowed")
	}
}

func stripPrefix(s, prefix string) (string, bool) {
	prefixLen := len(prefix)
	if len(s) <= prefixLen || !strings.HasPrefix(s, prefix) {
		return "", false
	}

	return s[prefixLen:], true
}
