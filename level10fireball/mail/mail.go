// mail
package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
)

type Email struct {
	To      string
	CC      string
	BCC     string
	Subject string
	Message string
	Sender  string
}

func sendMail(session *Session, email *Email) (successful bool) {
	// Connect to the remote SMTP server.
	mail_item, err := smtp.Dial(fmt.Sprintf("%s:%d", smtp_server, smtp_port))
	if err != nil {
		isErr(session, err, true, "sending e-mail", "sendMail")
		return false
	}

	// Set the sender and recipient.
	if email.Sender == "" {
		email.Sender = smtp_sender_address
	}
	mail_item.Mail(email.Sender)
	mail_item.Rcpt(email.To)
	// Send the email body
	w, err := mail_item.Data()
	if err != nil {
		isErr(session, err, true, "sending e-mail", "sendMail")
		return false
	}
	defer w.Close()
	buf := bytes.NewBufferString(email.Message)
	if _, err = buf.WriteTo(w); err != nil {
		isErr(session, err, true, "sending e-mail", "sendMail")
		return false
	}
	return true
}

/*
func main() {
        // Set up authentication information.
        auth := smtp.PlainAuth(
                "",
                "user@example.com",
                "password",
                "mail.example.com",
        )
        // Connect to the server, authenticate, set the sender and recipient,
        // and send the email all in one step.
        err := smtp.SendMail(
                "mail.example.com:25",
                auth,
                "sender@example.org",
                []string{"recipient@example.net"},
                []byte("This is the email body."),
        )
        if err != nil {
                log.Fatal(err)
        }
}
*/
