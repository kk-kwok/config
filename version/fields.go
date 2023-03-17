package version

var initialFields = map[string]interface{}{
	"service": ServiceName,
	"version": Version,
}

// InitialFields for zap InitialFields
func InitialFields() map[string]interface{} {
	return initialFields
}
