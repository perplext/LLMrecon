package notification

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

// EmailConfig represents the configuration for the email notification channel
type EmailConfig struct {
	SMTPServer   string
	SMTPPort     int
	Username     string
	Password     string
	FromAddress  string
	FromName     string
	ToAddresses  []string
	UseTLS       bool
	TemplatePath string
}

// EmailChannel represents a notification channel that sends notifications via email
type EmailChannel struct {
	id       string
	name     string
	config   EmailConfig
	template *template.Template
}

// NewEmailChannel creates a new email notification channel
func NewEmailChannel(config EmailConfig) (*EmailChannel, error) {
	var tmpl *template.Template
	var err error

	// Use default template if no template path is provided
	if config.TemplatePath == "" {
		tmpl, err = template.New("email").Parse(defaultEmailTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse default email template: %w", err)
		}
	} else {
		tmpl, err = template.ParseFiles(config.TemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse email template: %w", err)
		}
	}

	return &EmailChannel{
		id:       "email",
		name:     "Email",
		config:   config,
		template: tmpl,
	}, nil
}

// ID returns the unique identifier for the channel
func (e *EmailChannel) ID() string {
	return e.id
}

// Name returns the human-readable name of the channel
func (e *EmailChannel) Name() string {
	return e.name
}

// Deliver delivers a notification via email
func (e *EmailChannel) Deliver(notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	if len(e.config.ToAddresses) == 0 {
		return fmt.Errorf("no recipient email addresses configured")
	}

	// Prepare email data
	data := struct {
		Notification *Notification
		FromName     string
		CurrentTime  string
	}{
		Notification: notification,
		FromName:     e.config.FromName,
		CurrentTime:  time.Now().Format(time.RFC1123),
	}

	// Generate email body from template
	var bodyBuffer bytes.Buffer
	if err := e.template.Execute(&bodyBuffer, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Prepare email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", e.config.FromName, e.config.FromAddress)
	headers["To"] = strings.Join(e.config.ToAddresses, ", ")
	headers["Subject"] = fmt.Sprintf("[%s] %s", strings.ToUpper(string(notification.Severity)), notification.Title)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Construct message
	var message bytes.Buffer
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.Write(bodyBuffer.Bytes())

	// Connect to SMTP server
	var auth smtp.Auth
	if e.config.Username != "" {
		auth = smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPServer)
	}

	addr := fmt.Sprintf("%s:%d", e.config.SMTPServer, e.config.SMTPPort)

	// Send email
	var err error
	if e.config.UseTLS {
		// Create TLS config
		tlsConfig := &tls.Config{
			ServerName: e.config.SMTPServer,
		}

		// Connect to the server
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		// Create SMTP client
		client, err := smtp.NewClient(conn, e.config.SMTPServer)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		// Authenticate if needed
		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP authentication failed: %w", err)
			}
		}

		// Set the sender and recipients
		if err := client.Mail(e.config.FromAddress); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		for _, addr := range e.config.ToAddresses {
			if err := client.Rcpt(addr); err != nil {
				return fmt.Errorf("failed to set recipient %s: %w", addr, err)
			}
		}

		// Send the email body
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to open data connection: %w", err)
		}

		_, err = w.Write(message.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write email message: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close data connection: %w", err)
		}

		// Send the QUIT command and close the connection
		err = client.Quit()
		if err != nil {
			return fmt.Errorf("failed to quit SMTP connection: %w", err)
		}
	} else {
		// Use standard SMTP without TLS
		err = smtp.SendMail(
			addr,
			auth,
			e.config.FromAddress,
			e.config.ToAddresses,
			message.Bytes(),
		)
	}

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// CanDeliver checks if the channel can deliver the notification
func (e *EmailChannel) CanDeliver(notification *Notification) bool {
	// Email channel can only deliver if properly configured
	if len(e.config.ToAddresses) == 0 || e.config.FromAddress == "" {
		return false
	}

	// By default, only send critical notifications via email
	// unless the notification explicitly targets the email channel
	if notification.Severity != Critical {
		for _, channel := range notification.TargetChannels {
			if channel == e.id {
				return true
			}
		}
		return false
	}

	return true
}

// Default email template
const defaultEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Notification.Title}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: {{if eq .Notification.Severity "critical"}}#f44336{{else if eq .Notification.Severity "warning"}}#ff9800{{else}}#2196f3{{end}};
            color: white;
            padding: 10px 20px;
            border-radius: 5px 5px 0 0;
        }
        .content {
            padding: 20px;
            border: 1px solid #ddd;
            border-top: none;
            border-radius: 0 0 5px 5px;
        }
        .metadata {
            background-color: #f5f5f5;
            padding: 10px;
            margin-top: 20px;
            border-radius: 5px;
        }
        .action {
            margin-top: 20px;
            padding: 15px;
            background-color: #e8f5e9;
            border-radius: 5px;
        }
        .action-button {
            display: inline-block;
            padding: 10px 20px;
            background-color: #4caf50;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin-top: 10px;
        }
        .footer {
            margin-top: 20px;
            font-size: 12px;
            color: #777;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="header">
        <h2>{{.Notification.Title}}</h2>
    </div>
    <div class="content">
        <p>{{.Notification.Message}}</p>
        
        {{if .Notification.Metadata}}
        <div class="metadata">
            <h3>Additional Information</h3>
            <ul>
                {{range $key, $value := .Notification.Metadata}}
                <li><strong>{{$key}}:</strong> {{$value}}</li>
                {{end}}
            </ul>
        </div>
        {{end}}
        
        {{if .Notification.RequiresAction}}
        <div class="action">
            <h3>Action Required</h3>
            <p>{{if .Notification.ActionLabel}}{{.Notification.ActionLabel}}{{else}}Please take action{{end}}</p>
            {{if .Notification.ActionURL}}
            <a href="{{.Notification.ActionURL}}" class="action-button">Take Action</a>
            {{end}}
        </div>
        {{end}}
    </div>
    <div class="footer">
        <p>This notification was sent by {{.FromName}} at {{.CurrentTime}}</p>
    </div>
</body>
</html>`
