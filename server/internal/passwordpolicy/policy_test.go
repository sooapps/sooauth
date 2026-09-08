package passwordpolicy

import "testing"

func TestPolicyValidate(t *testing.T) {
	p := Policy{MinLength: 8, RequireUppercase: true}
	if err := p.Validate("short"); err == nil {
		t.Fatal("expected min length error")
	}
	if err := p.Validate("alllower1"); err == nil {
		t.Fatal("expected uppercase error")
	}
	if err := p.Validate("Goodpass1"); err != nil {
		t.Fatalf("expected valid password, got %v", err)
	}
}
