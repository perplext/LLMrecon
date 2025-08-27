// Package converters provides conversion functions between different types
package converters

// GenericConverter represents a generic model converter
type GenericConverter struct{}

// NewUserConverter creates a new user converter
func NewUserConverter() *GenericConverter {
	return &GenericConverter{}
}

// NewSessionConverter creates a new session converter
func NewSessionConverter() *GenericConverter {
	return &GenericConverter{}
}

// NewSecurityConverter creates a new security converter
func NewSecurityConverter() *GenericConverter {
	return &GenericConverter{}
}

// NewIncidentConverter creates a new incident converter
func NewIncidentConverter() *GenericConverter {
	return &GenericConverter{}
}

// NewVulnerabilityConverter creates a new vulnerability converter
func NewVulnerabilityConverter() *GenericConverter {
	return &GenericConverter{}
}

// NewAuditLogConverter creates a new audit log converter
func NewAuditLogConverter() *GenericConverter {
	return &GenericConverter{}
}
