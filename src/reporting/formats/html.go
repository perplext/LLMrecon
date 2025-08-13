package formats

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/perplext/LLMrecon/src/reporting/api"
)

// HTMLFormatter is a formatter for HTML reports
type HTMLFormatter struct {
	// template is the HTML template for reports
	template *template.Template
}

// NewHTMLFormatter creates a new HTML formatter with the default template
func NewHTMLFormatter() (*HTMLFormatter, error) {
	// Parse default template
	tmpl, err := template.New("report").Parse(defaultHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse default HTML template: %w", err)
	}

	return &HTMLFormatter{
		template: tmpl,
	}, nil
}

// NewHTMLFormatterWithTemplate creates a new HTML formatter with a custom template
func NewHTMLFormatterWithTemplate(templatePath string) (*HTMLFormatter, error) {
	// Read template file
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template from file %s: %w", templatePath, err)
	}

	return &HTMLFormatter{
		template: tmpl,
	}, nil
}

// FormatReport formats a report and writes it to the given writer
func (f *HTMLFormatter) FormatReport(results api.TestResults, writer io.Writer) error {
	// Create a simple HTML report
	data := struct {
		Results     api.TestResults
		GeneratedAt time.Time
		Title       string
		FormatTime  func(time.Time) string
		CurrentYear int
	}{
		Results:     results,
		GeneratedAt: time.Now(),
		Title:       "Test Results Report",
		FormatTime: func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		CurrentYear: time.Now().Year(),
	}

	// Execute template
	return f.template.Execute(writer, data)
}

// Format formats a report as HTML
func (f *HTMLFormatter) Format(ctx context.Context, reportInterface interface{}, optionsInterface interface{}) ([]byte, error) {
	results, ok := reportInterface.(api.TestResults)
	if !ok {
		return nil, fmt.Errorf("expected api.TestResults, got %T", reportInterface)
	}
	
	// Create a buffer to hold the HTML data
	buf := &bytes.Buffer{}
	
	// Use the FormatReport method to write to the buffer
	err := f.FormatReport(results, buf)
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// GetFormat returns the format supported by this formatter
func (f *HTMLFormatter) GetFormat() api.ReportFormat {
	return api.HTMLFormat
}

// WriteToFile writes a report to a file
func (f *HTMLFormatter) WriteToFile(ctx context.Context, reportInterface interface{}, optionsInterface interface{}, filePath string) error {
	results, ok := reportInterface.(api.TestResults)
	if !ok {
		return fmt.Errorf("expected api.TestResults, got %T", reportInterface)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Format and write the report
	if err := f.FormatReport(results, file); err != nil {
		return fmt.Errorf("failed to write report to file %s: %w", filePath, err)
	}

	return nil
}

// defaultHTMLTemplate is the default HTML template for reports
var defaultHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        :root {
            --primary-color: #3498db;
            --secondary-color: #2c3e50;
            --success-color: #2ecc71;
            --warning-color: #f39c12;
            --danger-color: #e74c3c;
            --info-color: #1abc9c;
            --light-color: #ecf0f1;
            --dark-color: #34495e;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: var(--secondary-color);
            background-color: #f8f9fa;
            margin: 0;
            padding: 0;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        header {
            background-color: var(--light-color);
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
            border-left: 5px solid var(--primary-color);
        }
        
        h1, h2, h3, h4 {
            margin-top: 0;
            color: var(--secondary-color);
        }
        
        h1 {
            font-size: 2.2rem;
        }
        
        h2 {
            font-size: 1.8rem;
            border-bottom: 2px solid var(--light-color);
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        
        h3 {
            font-size: 1.4rem;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 20px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        
        th, td {
            padding: 12px 15px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        
        th {
            background-color: var(--light-color);
            color: var(--secondary-color);
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>{{.Title}}</h1>
            <p>Generated: {{.FormatTime .GeneratedAt}}</p>
        </header>
        
        <section>
            <h2>Test Results</h2>
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Name</th>
                        <th>Severity</th>
                        <th>Status</th>
                        <th>Category</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Results}}
                    <tr>
                        <td>{{.ID}}</td>
                        <td>{{.Name}}</td>
                        <td>{{.Severity}}</td>
                        <td>{{.Status}}</td>
                        <td>{{.Category}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>
        
        <footer>
            <p>LLMrecon Tool &copy; {{.CurrentYear}}</p>
        </footer>
    </div>
</body>
</html>`
