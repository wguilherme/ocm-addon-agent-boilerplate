package errors

import "fmt"

type AnalyzerErrorType string

const (
	ErrorTypeCollectionFailed   AnalyzerErrorType = "CollectionFailed"
	ErrorTypeProcessingFailed   AnalyzerErrorType = "ProcessingFailed"
	ErrorTypeTransmissionFailed AnalyzerErrorType = "TransmissionFailed"
	ErrorTypeValidationFailed   AnalyzerErrorType = "ValidationFailed"
)

type CollectionError struct {
	AnalyzerName string
	Cause        error
}

func (e *CollectionError) Error() string {
	return fmt.Sprintf("collection failed in analyzer '%s': %v", e.AnalyzerName, e.Cause)
}

func (e *CollectionError) Unwrap() error {
	return e.Cause
}

type ProcessingError struct {
	ProcessorName string
	Cause         error
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("processing failed in processor '%s': %v", e.ProcessorName, e.Cause)
}

func (e *ProcessingError) Unwrap() error {
	return e.Cause
}

type TransmissionError struct {
	TransmitterName string
	Cause           error
}

func (e *TransmissionError) Error() string {
	return fmt.Sprintf("transmission failed in transmitter '%s': %v", e.TransmitterName, e.Cause)
}

func (e *TransmissionError) Unwrap() error {
	return e.Cause
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

var (
	ErrNilConfig = &ValidationError{
		Field:   "config",
		Message: "SyncConfig cannot be nil",
	}

	ErrNilSpokeClient = &ValidationError{
		Field:   "config.SpokeClient",
		Message: "SpokeClient cannot be nil",
	}

	ErrNilHubClient = &ValidationError{
		Field:   "config.HubClient",
		Message: "HubClient cannot be nil",
	}

	ErrEmptyClusterName = &ValidationError{
		Field:   "config.SpokeClusterName",
		Message: "SpokeClusterName cannot be empty",
	}
)

func NewCollectionError(analyzerName string, cause error) error {
	return &CollectionError{
		AnalyzerName: analyzerName,
		Cause:        cause,
	}
}

func NewProcessingError(processorName string, cause error) error {
	return &ProcessingError{
		ProcessorName: processorName,
		Cause:         cause,
	}
}

func NewTransmissionError(transmitterName string, cause error) error {
	return &TransmissionError{
		TransmitterName: transmitterName,
		Cause:           cause,
	}
}
