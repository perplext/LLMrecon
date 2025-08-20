package formats

import (
	"github.com/perplext/LLMrecon/src/reporting/api"
)

// init registers all formatters with the API registry
func init() {
	// Register JSON formatter
	api.RegisterFormatter(api.JSONFormat, func(options map[string]interface{}) (api.ReportFormatter, error) {
		pretty := true
		if val, ok := options["pretty"]; ok {
			if prettyBool, ok := val.(bool); ok {
				pretty = prettyBool
			}
		}
		return NewJSONFormatter(pretty), nil
	})
	
	// Register JSONL formatter
	api.RegisterFormatter(api.JSONLFormat, func(options map[string]interface{}) (api.ReportFormatter, error) {
		return NewJSONLFormatter(), nil
	})
	
	// Register CSV formatter
	api.RegisterFormatter(api.CSVFormat, func(options map[string]interface{}) (api.ReportFormatter, error) {
		delimiter := ','
		if val, ok := options["delimiter"]; ok {
			if delimiterRune, ok := val.(rune); ok {
				delimiter = delimiterRune
			} else if delimiterStr, ok := val.(string); ok && len(delimiterStr) > 0 {
				delimiter = rune(delimiterStr[0])
			}
		}
		
		includeHeaders := true
		if val, ok := options["include_headers"]; ok {
			if includeHeadersBool, ok := val.(bool); ok {
				includeHeaders = includeHeadersBool
			}
		}
		
		return NewCSVFormatter(delimiter, includeHeaders), nil
	})
	
	// Register Excel formatter
	api.RegisterFormatter(api.ExcelFormat, func(options map[string]interface{}) (api.ReportFormatter, error) {
		includeRawData := false
		if val, ok := options["include_raw_data"]; ok {
			if includeRawDataBool, ok := val.(bool); ok {
				includeRawData = includeRawDataBool
			}
		}
		
		return NewExcelFormatter(includeRawData), nil
	})
	
	// Register Text formatter
	api.RegisterFormatter(api.TextFormat, func(options map[string]interface{}) (api.ReportFormatter, error) {
		detailed := false
		if val, ok := options["detailed"]; ok {
			if detailedBool, ok := val.(bool); ok {
				detailed = detailedBool
			}
		}
		
		return NewTextFormatter(detailed), nil
	})
	
	// Register Markdown formatter
	api.RegisterFormatter(api.MarkdownFormat, func(options map[string]interface{}) (api.ReportFormatter, error) {
		includeRawData := false
		if val, ok := options["include_raw_data"]; ok {
			if includeRawDataBool, ok := val.(bool); ok {
				includeRawData = includeRawDataBool
			}
		}
		
		return NewMarkdownFormatter(includeRawData), nil
	})
	
	// Register PDF formatter
	api.RegisterFormatter(api.PDFFormat, func(options map[string]interface{}) (api.ReportFormatter, error) {
		customTemplate := ""
		if val, ok := options["template_path"]; ok {
			if templatePath, ok := val.(string); ok {
				customTemplate = templatePath
			}
		}
		
		return NewPDFFormatter(customTemplate), nil
	})
	
	// Register HTML formatter
	api.RegisterFormatter(api.HTMLFormat, func(options map[string]interface{}) (api.ReportFormatter, error) {
		customTemplate := ""
		if val, ok := options["template_path"]; ok {
			if templatePath, ok := val.(string); ok {
				customTemplate = templatePath
				formatter, err := NewHTMLFormatterWithTemplate(customTemplate)
				return formatter, err
			}
		}
		
		formatter, err := NewHTMLFormatter()
		return formatter, err
	})
}
