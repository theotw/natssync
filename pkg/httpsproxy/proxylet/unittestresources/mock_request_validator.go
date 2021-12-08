package unittestresources

type mockRequestValidator struct {
	ValidationError error
}

func NewMockRequestValidator() *mockRequestValidator {
	return &mockRequestValidator{
		ValidationError: nil,
	}
}

func (m *mockRequestValidator) IsValidRequest(string) error {
	return m.ValidationError
}
