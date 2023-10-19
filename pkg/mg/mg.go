package mg

import "github.com/mailgun/mailgun-go/v4"

type Mailgun struct {
	Domain string
	ApiKey string

	mg *mailgun.MailgunImpl
}

func (m *Mailgun) Init() {
	// Create an instance of the Mailgun Client
	mg := mailgun.NewMailgun(m.Domain, m.ApiKey)
	mg.SetAPIBase(mailgun.APIBaseEU)

	m.mg = mg
}
