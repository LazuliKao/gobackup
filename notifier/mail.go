package notifier

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/smtp"
	"sort"
	"strings"
)

// TLS modes for mail notifier
const (
	TLSModeNone     = "none"     // No encryption at all
	TLSModeStartTLS = "starttls" // STARTTLS upgrade
	TLSModeTLS      = "tls"      // Full TLS from start
)

type Mail struct {
	// Base is the base notifier
	from     string
	to       []string
	username string
	password string
	host     string
	port     string
	tls      string // "none", "starttls", or "tls"
}

func NewMail(base *Base) (*Mail, error) {
	base.viper.SetDefault("port", "25")

	username := base.viper.GetString("username")
	if len(username) == 0 {
		return nil, fmt.Errorf("username is required for mail notifier")
	}

	from := base.viper.GetString("from")
	if len(from) == 0 {
		from = username
	}

	// Determine TLS mode from "tls" config
	// Supports: "none", "starttls", "tls", true (=tls), false (=starttls)
	tlsMode := TLSModeStartTLS // default
	tlsValue := base.viper.Get("tls")
	if tlsValue != nil {
		switch v := tlsValue.(type) {
		case string:
			tlsMode = v
		case bool:
			if v {
				tlsMode = TLSModeTLS
			} else {
				tlsMode = TLSModeStartTLS
			}
		}
	}

	return &Mail{
		username: username,
		password: base.viper.GetString("password"),
		to:       strings.Split(base.viper.GetString("to"), ","),
		from:     from,
		host:     base.viper.GetString("host"),
		port:     base.viper.GetString("port"),
		tls:      tlsMode,
	}, nil
}

func (s Mail) getAddr() string {
	return fmt.Sprintf("%s:%s", s.host, s.port)
}

func (s Mail) getAuth() smtp.Auth {
	return smtp.PlainAuth("", s.username, s.password, s.host)
}

// unencryptedAuth is a custom Auth that works over unencrypted connections
type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// Allow unencrypted connections by setting TLS to true temporarily
	server.TLS = true
	return a.Auth.Start(server)
}

func (s Mail) getUnencryptedAuth() smtp.Auth {
	return unencryptedAuth{smtp.PlainAuth("", s.username, s.password, s.host)}
}

func (s Mail) buildBody(title string, message string) string {
	headers := make(map[string]string)
	headers["From"] = s.from
	headers["To"] = strings.Join(s.to, ",")
	headers["Subject"] = title
	headers["Content-Type"] = `text/plain; charset="utf-8"`
	headers["Content-Transfer-Encoding"] = "base64"

	// Sort headers by key
	var keys []string
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headerTexts := []string{}

	for _, k := range keys {
		headerTexts = append(headerTexts, fmt.Sprintf("%s: %s", k, headers[k]))
	}

	return fmt.Sprintf("%s\n%s", strings.Join(headerTexts, "\n"), base64.StdEncoding.EncodeToString([]byte(message)))
}

func (s *Mail) notify(title string, message string) error {
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	msg := s.buildBody(title, message)

	switch s.tls {
	case TLSModeTLS:
		return s.sendByTLS([]byte(msg))
	case TLSModeNone:
		return s.sendPlain([]byte(msg))
	case TLSModeStartTLS:
		fallthrough
	default:
		return s.sendByStartTLS([]byte(msg))
	}
}

// sendPlain sends email without any encryption (TLS mode: none)
func (s *Mail) sendPlain(msg []byte) error {
	conn, err := smtp.Dial(s.getAddr())
	if err != nil {
		return err
	}
	defer conn.Close()

	if len(s.password) > 0 {
		ok, _ := conn.Extension("AUTH")
		if ok {
			// Use unencrypted auth for plain connections
			auth := s.getUnencryptedAuth()
			if err = conn.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err = conn.Mail(s.from); err != nil {
		return err
	}
	for _, addr := range s.to {
		if err = conn.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := conn.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	conn.Quit()
	return nil
}

// sendByStartTLS sends email using STARTTLS upgrade (TLS mode: starttls)
func (s *Mail) sendByStartTLS(msg []byte) error {
	conn, err := smtp.Dial(s.getAddr())
	if err != nil {
		return err
	}
	defer conn.Close()

	// Upgrade to TLS using STARTTLS
	if ok, _ := conn.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: s.host}
		if err = conn.StartTLS(config); err != nil {
			return err
		}
	}

	if len(s.password) > 0 {
		ok, _ := conn.Extension("AUTH")
		if ok {
			auth := s.getAuth()
			if err = conn.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err = conn.Mail(s.from); err != nil {
		return err
	}
	for _, addr := range s.to {
		if err = conn.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := conn.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	conn.Quit()
	return nil
}

func (s *Mail) sendByTLS(msg []byte) error {
	conn, err := tls.Dial("tcp", s.getAddr(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}

	if len(s.password) > 0 {
		ok, _ := client.Extension("AUTH")
		if !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		auth := s.getAuth()
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(s.from); err != nil {
		return err
	}
	for _, addr := range s.to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	client.Quit()
	return nil
}
