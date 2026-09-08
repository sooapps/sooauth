package mail

import (
	"testing"
)

func TestParseSMTPURL_emailUsernamePort465(t *testing.T) {
	addr, auth, tls, err := parseSMTPURL("smtp://info%40sooapps.com:secret@promail.internetbilisim.net:465")
	if err != nil {
		t.Fatal(err)
	}
	if addr != "promail.internetbilisim.net:465" {
		t.Fatalf("addr = %q", addr)
	}
	if !tls {
		t.Fatal("expected implicit TLS for port 465")
	}
	if auth == nil {
		t.Fatal("expected auth")
	}
}

func TestParseSMTPURL_specialPasswordMustBeEncoded(t *testing.T) {
	_, _, _, err := parseSMTPURL("smtp://info%40sooapps.com:20032002yu?!@promail.internetbilisim.net:465")
	if err == nil {
		t.Fatal("expected parse error for unencoded ? in password")
	}
}

func TestParseSMTPURL_encodedSpecialPassword(t *testing.T) {
	addr, auth, tls, err := parseSMTPURL("smtp://info%40sooapps.com:20032002yu%3F%21@promail.internetbilisim.net:465")
	if err != nil {
		t.Fatal(err)
	}
	if addr != "promail.internetbilisim.net:465" || !tls || auth == nil {
		t.Fatalf("addr=%q tls=%v auth=%v", addr, tls, auth)
	}
}

func TestParseSMTPParts_specialPassword(t *testing.T) {
	addr, auth, tls, err := parseSMTPParts("promail.internetbilisim.net", "465", "info@sooapps.com", "20032002yu?!")
	if err != nil {
		t.Fatal(err)
	}
	if addr != "promail.internetbilisim.net:465" || !tls || auth == nil {
		t.Fatalf("addr=%q tls=%v auth=%v", addr, tls, auth)
	}
}

func TestParseSMTPURL_simpleMailhog(t *testing.T) {
	addr, auth, tls, err := parseSMTPURL("smtp://user:pass@localhost:1025")
	if err != nil {
		t.Fatal(err)
	}
	if addr != "localhost:1025" || tls {
		t.Fatalf("addr=%q tls=%v", addr, tls)
	}
	if auth == nil {
		t.Fatal("expected auth")
	}
}
