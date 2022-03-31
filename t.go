package main

// Mocker ...
type Mocker interface {
	Mock(input int, output int) (code int, str string)
}

// FakeMocker ...
type FakeMocker struct {
	MockName string
}

// Mock ...
func (fm *FakeMocker) Mock(input int, output int) (code int, str string) {
	return 0, ""
}

// Test ...
type Test struct {
	embed *Manager
}

// Source defines enum of Source values
type Source string

// express source types.
const (
	// Operator ...
	Operator Source = "OPERATOR"
	// IOS ...
	IOS Source = "IOS"
	// Android ...
	Android Source = "ANDROID"
	// Partner ...
	Partner Source = "PARTNER"
	// GEWEB ...
	GEWEB Source = "GEWEB"
	// WEBAPI ...
	WEBAPI Source = "WEBAPI" // source of a DeliveryLinkRequest if it was created via web-microsite

	defaultBookingCodePrefix = "EXPRESS"
)
