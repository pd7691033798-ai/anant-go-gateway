package featurephone

import (
	"sync"
)

type ContactFilter struct {
	mu          sync.RWMutex
	familyNodes map[string]string // phone -> name/relation
}

func NewContactFilter() *ContactFilter {
	cf := &ContactFilter{
		familyNodes: make(map[string]string),
	}
	// डिफ़ॉल्ट एडमिन व परिवार के नंबर
	cf.familyNodes["9024414973"] = "एडमिन"
	return cf
}

func (cf *ContactFilter) AddPersonalContact(phone, label string) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.familyNodes[phone] = label
}

func (cf *ContactFilter) IsPersonalContact(phone string) (bool, string) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	label, exists := cf.familyNodes[phone]
	return exists, label
}
