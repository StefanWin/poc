package api

type ConversionProcessingResponse struct {
	ConversionID string `json:"conversionId"`
}

type ConversionResult struct {
	ConversionID string `json:"conversionId"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	ResultFile   []byte `json:"resultFile"`
}

type ConversionStatusResponse struct {
	Status string            `json:"status"`
	Result *ConversionResult `json:"result"`
}

type ConversionRequest struct {
	ConversionRequestBody ConversionRequestBody
	ConversionStatus      string
	ExternalConversionID  string
	WorkerConversionID    string // TODO : make ptr so it can be nil
}

type ConversionRequestBody struct {
	File           []byte `json:"file"` // TODO : remove since its not needed
	Filename       string `json:"filename"`
	OriginalFormat string `json:"originalFormat"` // TODO : make ptr so it can be nil
	TargetFormat   string `json:"targetFormat"`
}

type ConversionStatus struct {
	ConversionID string `json:"conversionId"`
	Status       string `json:"status"`
	Path         string `json:"path"`
	Retries      int    `json:"retries"`
	SourceFormat string `json:"sourceFormat"`
	TargetFormat string `json:"targetFormat"`
}

type ConversionQueueStatus struct {
	Conversions         []ConversionStatus `json:"conversions"`
	RemainingConversion int                `json:"remainingConversion"`
}
