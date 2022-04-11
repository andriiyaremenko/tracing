package tracing

import "net/http"

var (
	// Metadata reader from Header using default Header names.
	DefaultMetadataReadHeader = MetadataReadHeader(
		HeaderRequestID,
		HeaderCausationID,
		HeaderCorrelationID,
	)
	// Metadata writer to Header using default Header names.
	DefaultMetadataWriteHeader = MetadataWriteHeader(
		HeaderRequestID,
		HeaderCausationID,
		HeaderCorrelationID,
	)
	// Metadata options with default Header names.
	DefaultMetadataOptions = MetadataOptionsWithHeader(
		HeaderRequestID,
		HeaderCausationID,
		HeaderCorrelationID,
	)
)

// Metadata carries additional information not used in command execution.
type Metadata struct {
	// Unique event ID.
	ID string
	// Root event ID that triggered execution of the program.
	CorrelationID string
	// ID of event that caused execution of current event.
	CausationID string
}

// Creates new Metadata.
func NewMetadata(id string) Metadata {
	return Metadata{
		ID:            id,
		CorrelationID: id,
		CausationID:   id,
	}
}

// New Metadata for next event in execution chain.
func NextMetadata(m Metadata, id string) Metadata {
	return Metadata{
		ID:            id,
		CausationID:   m.ID,
		CorrelationID: m.CorrelationID,
	}
}

// Checks if Metadata is valid.
func ValidMetadata(m *Metadata) bool {
	return m.ID != "" && m.CausationID != "" && m.CorrelationID != ""
}

// Metadata reader from Header using provided Header names.
// Will canonicalize provided names.
func MetadataReadHeader(
	requestID, causationID, correlationID string,
) func(header http.Header, id string) (Metadata, bool) {
	return func(header http.Header, id string) (Metadata, bool) {
		m := Metadata{
			ID:            header.Get(requestID),
			CorrelationID: header.Get(correlationID),
			CausationID:   header.Get(causationID),
		}

		if ValidMetadata(&m) {
			return m, true
		}

		return NewMetadata(id), false
	}
}

// Metadata writer to Header using provided Header names.
// Will canonicalize provided names.
func MetadataWriteHeader(
	requestID, causationID, correlationID string,
) func(http.Header, Metadata) {
	return func(header http.Header, m Metadata) {
		header.Set(requestID, m.ID)
		header.Set(causationID, m.CausationID)
		header.Set(correlationID, m.CorrelationID)
	}
}

// Metadata options with provided Header reader and writer.
func MetadataOptions(
	read func(http.Header, string) (Metadata, bool),
	write func(http.Header, Metadata),
) Options[Metadata] {
	return func() (ReadHeader[Metadata], WriteHeader[Metadata], Next[Metadata]) {
		return read, write, NextMetadata
	}
}

// Metadata options with provided Header names.
func MetadataOptionsWithHeader(requestID, causationID, correlationID string) Options[Metadata] {
	return MetadataOptions(
		MetadataReadHeader(requestID, causationID, correlationID),
		MetadataWriteHeader(requestID, causationID, correlationID),
	)
}
