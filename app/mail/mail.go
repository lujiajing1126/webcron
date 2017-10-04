package mail

import (
	"github.com/astaxie/beego"
	"time"
	"gopkg.in/gomail.v2"
)

var (
	dialer *gomail.Dialer
	sendCh chan *gomail.Message
	from   string
)

func init() {
	queueSize, _ := beego.AppConfig.Int("mail.queue_size")
	host := beego.AppConfig.String("mail.host")
	port, _ := beego.AppConfig.Int("mail.port")
	username := beego.AppConfig.String("mail.user")
	password := beego.AppConfig.String("mail.password")
	from = beego.AppConfig.String("mail.from")
	if port == 0 {
		port = 25
	}
	dialer = gomail.NewDialer(host, port, username, password)

	sendCh = make(chan *gomail.Message, queueSize)

	go func() {
		for {
			select {
			case m, ok := <-sendCh:
				if !ok {
					return
				}
				if err := dialer.DialAndSend(m); err != nil {
					beego.Error("SendMail:", err.Error())
				}
			}
		}
	}()
}

func SendMail(address, name, subject, content string, cc []string) bool {
	message := gomail.NewMessage()
	message.SetHeader("From", from)
	message.SetHeader("To", address)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", content)
	if len(cc) > 0 {
		message.SetHeader("Cc", cc...)
	}

	select {
	case sendCh <- message:
		return true
	case <-time.After(time.Second * 3):
		return false
	}
}
