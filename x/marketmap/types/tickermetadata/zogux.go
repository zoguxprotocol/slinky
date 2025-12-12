package tickermetadata

import "encoding/json"

// Zogux is the Ticker.Metadata_JSON published to every Ticker in the x/marketmap module on Zogux.
type Zogux struct {
	// ReferencePrice gives a spot price for that Ticker at the point in time when the ReferencePrice was updated.
	// You should _not_ use this for up-to-date/instantaneous spot pricing data since it is updated infrequently.
	// The price is scaled by Ticker.Decimals.
	ReferencePrice uint64 `json:"reference_price"`
	// Liquidity gives a _rough_ estimate of the amount of liquidity in the Providers for a given Market.
	// It is _not_ updated in coordination with spot prices and only gives rough order of magnitude accuracy at the time
	// which the update for it is published.
	// The liquidity value stored here is USD denominated.
	Liquidity uint64 `json:"liquidity"`
	// AggregateIDs contains a list of AggregatorIDs associated with the ticker.
	// This field may not be populated if no aggregator currently indexes this Ticker.
	AggregateIDs []AggregatorID `json:"aggregate_ids"`
	// CrossLaunch is an optional bool that indicates whether this ticker should be
	// launched as a cross-margin market (instead of isolated margin).
	// If omitted, it is set to false by default.
	CrossLaunch bool `json:"cross_launch,omitempty"`
}

// NewZogux returns a new Zogux instance.
func NewZogux(referencePrice, liquidity uint64, aggregateIDs []AggregatorID, crossLaunch bool) Zogux {
	return Zogux{
		ReferencePrice: referencePrice,
		Liquidity:      liquidity,
		AggregateIDs:   aggregateIDs,
		CrossLaunch:    crossLaunch,
	}
}

// MarshalZogux returns the JSON byte encoding of the Zogux.
func MarshalZogux(m Zogux) ([]byte, error) {
	return json.Marshal(m)
}

// ZoguxFromJSONString returns a Zogux instance from a JSON string.
func ZoguxFromJSONString(jsonString string) (Zogux, error) {
	var elem Zogux
	err := json.Unmarshal([]byte(jsonString), &elem)
	return elem, err
}

// ZoguxFromJSONBytes returns a Zogux instance from JSON bytes.
func ZoguxFromJSONBytes(jsonBytes []byte) (Zogux, error) {
	var elem Zogux
	err := json.Unmarshal(jsonBytes, &elem)
	return elem, err
}
