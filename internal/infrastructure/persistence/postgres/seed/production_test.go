package seed

import "testing"

func TestValidateProductionCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		email   string
		pass    string
		first   string
		last    string
		wantErr bool
	}{
		{name: "valid", email: "admin@rehil.gallery", pass: "SecurePass1", first: "Admin", last: "User"},
		{name: "missing email", email: "", pass: "SecurePass1", first: "Admin", last: "User", wantErr: true},
		{name: "short password", email: "admin@rehil.gallery", pass: "short1", first: "Admin", last: "User", wantErr: true},
		{name: "no digit", email: "admin@rehil.gallery", pass: "NoDigitsHere", first: "Admin", last: "User", wantErr: true},
		{name: "no letter", email: "admin@rehil.gallery", pass: "12345678", first: "Admin", last: "User", wantErr: true},
		{name: "missing name", email: "admin@rehil.gallery", pass: "SecurePass1", first: "", last: "User", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateProductionCredentials(tc.email, tc.pass, tc.first, tc.last)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateProductionCredentials() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestProductionOptionsNormalizedStaffPair(t *testing.T) {
	t.Parallel()

	_, err := ProductionOptions{
		AdminEmail:    "admin@rehil.gallery",
		AdminPassword: "SecurePass1",
		StaffEmail:    "staff@rehil.gallery",
	}.normalized()
	if err == nil {
		t.Fatal("expected error when staff password is missing")
	}
}
