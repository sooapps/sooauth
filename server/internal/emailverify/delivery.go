package emailverify

import "strings"

type Delivery string

const (
	DeliveryLink Delivery = "link"
	DeliveryCode Delivery = "code"
	DeliveryBoth Delivery = "both"
)

func ParseDelivery(raw string) Delivery {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(DeliveryCode):
		return DeliveryCode
	case string(DeliveryBoth):
		return DeliveryBoth
	default:
		return DeliveryLink
	}
}

func (d Delivery) String() string {
	return string(d)
}

func (d Delivery) IncludesLink() bool {
	return d == DeliveryLink || d == DeliveryBoth
}

func (d Delivery) IncludesCode() bool {
	return d == DeliveryCode || d == DeliveryBoth
}

func (d Delivery) Valid() bool {
	return d == DeliveryLink || d == DeliveryCode || d == DeliveryBoth
}
