package cmd

// Test ...
type Test struct {
	a int
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
