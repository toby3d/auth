package jwk

import "fmt"

func (k KeyUsageType) String() string {
	return string(k)
}

func (k *KeyUsageType) Accept(v interface{}) error {
	switch v := v.(type) {
	case KeyUsageType:
		switch v {
		case ForSignature, ForEncryption:
			*k = v
			return nil
		default:
			return fmt.Errorf("invalid key usage type %s", v)
		}
	case string:
		switch v {
		case ForSignature.String(), ForEncryption.String():
			*k = KeyUsageType(v)
			return nil
		default:
			return fmt.Errorf("invalid key usage type %s", v)
		}
	}

	return fmt.Errorf("invalid value for key usage type %s", v)
}
