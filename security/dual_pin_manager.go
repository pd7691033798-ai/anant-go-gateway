package security

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type DualPinVault struct {
	HashedChildPin  string
	HashedParentPin string
	IsSessionLocked bool
}

func hashPin(pin string) string {
	sum := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(sum[:])
}

func NewDualPinVault(childPin, parentPin string) (*DualPinVault, error) {
	if len(childPin) != 4 || len(parentPin) != 4 {
		return nil, errors.New("पिन ठीक 4 अंक का होना चाहिए")
	}
	if childPin == parentPin {
		return nil, errors.New("चाइल्ड और पैरेंट पिन अलग होने चाहिए")
	}
	return &DualPinVault{
		HashedChildPin:  hashPin(childPin),
		HashedParentPin: hashPin(parentPin),
		IsSessionLocked: false,
	}, nil
}

func (v *DualPinVault) VerifyChild(pin string) bool {
	return hashPin(pin) == v.HashedChildPin && !v.IsSessionLocked
}

func (v *DualPinVault) VerifyParentOverride(pin string) bool {
	if hashPin(pin) == v.HashedParentPin {
		v.IsSessionLocked = false
		return true
	}
	return false
}

func (v *DualPinVault) LockByAntiCheat() {
	v.IsSessionLocked = true
}
