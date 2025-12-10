package imap

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/emersion/go-imap/client"
)

// ConnectAndLogin connects to an IMAP server and logs in
func ConnectAndLogin(server string, port int, email, password string) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", server, port)
	log.Printf("Connecting to IMAP server: %s", addr)

	// Connect to server
	c, err := client.DialTLS(addr, nil)
	if err != nil {
		// Try non-TLS if TLS fails, though usually 993 is TLS
		log.Printf("TLS connection failed, trying plain/STARTTLS: %v", err)
		c, err = client.Dial(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
		}
        
        // Check if STARTTLS is supported and use it if possible
        if ok, _ := c.SupportStartTLS(); ok {
             if err := c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
                 return nil, fmt.Errorf("failed to start TLS: %w", err)
             }
        }
	}

	log.Println("Connected to IMAP server")

	// Login
	if err := c.Login(email, password); err != nil {
		return nil, fmt.Errorf("failed to login to IMAP server: %w", err)
	}

	log.Println("Logged in to IMAP server")
	return c, nil
}
