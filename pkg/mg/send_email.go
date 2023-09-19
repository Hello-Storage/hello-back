package mg

import (
	"context"
	"time"
)

// Send email by using mailgun
func (m *Mailgun) SendEmail(
	from string,
	to string,
	subject string,
	template string,
	params map[string]interface{},
) (string, error) {
	message := m.mg.NewMessage(from, subject, "", to)

	message.SetTemplate(template)
	for k, v := range params {
		message.AddVariable(k, v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, id, err := m.mg.Send(ctx, message)

	return id, err
}
