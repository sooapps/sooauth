package emailverify

import "testing"

func TestParseDelivery(t *testing.T) {
	if ParseDelivery("code") != DeliveryCode {
		t.Fatal("expected code")
	}
	if ParseDelivery("both") != DeliveryBoth {
		t.Fatal("expected both")
	}
	if ParseDelivery("") != DeliveryLink {
		t.Fatal("expected default link")
	}
}

func TestGenerateCode(t *testing.T) {
	plain, hash, err := GenerateCode()
	if err != nil {
		t.Fatal(err)
	}
	if !ValidCode(plain) {
		t.Fatalf("invalid code %q", plain)
	}
	if hash == "" || hash == plain {
		t.Fatal("expected hashed code")
	}
}
