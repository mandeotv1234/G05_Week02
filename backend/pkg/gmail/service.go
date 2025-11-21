package gmail

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	emaildomain "ga03-backend/internal/email/domain"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// TokenUpdateFunc is a callback function that handles token updates
type TokenUpdateFunc = emaildomain.TokenUpdateFunc

type Service struct {
	clientID     string
	clientSecret string
}

type notifyTokenSource struct {
	src      oauth2.TokenSource
	current  *oauth2.Token
	callback TokenUpdateFunc
}

func (s *notifyTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.src.Token()
	if err != nil {
		return nil, err
	}
	if s.callback != nil && s.current.AccessToken != t.AccessToken {
		s.current = t
		// Execute callback in background to not block the request? 
		// Better to block to ensure consistency, or at least log error.
		if err := s.callback(t); err != nil {
			fmt.Printf("Failed to update token: %v\n", err)
		}
	}
	return t, nil
}

func NewService(clientID, clientSecret string) *Service {
	return &Service{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// GetGmailService creates Gmail service with user's access token
func (s *Service) GetGmailService(ctx context.Context, accessToken, refreshToken string, onTokenRefresh TokenUpdateFunc) (*gmail.Service, error) {
	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}

	// Only force refresh if we have a refresh token
	if refreshToken != "" {
		token.Expiry = time.Now()
	}

	config := &oauth2.Config{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		Endpoint:     google.Endpoint,
	}

	tokenSource := config.TokenSource(ctx, token)
	
	// Wrap token source to detect refreshes
	wrappedSource := &notifyTokenSource{
		src:      tokenSource,
		current:  token,
		callback: onTokenRefresh,
	}

	client := oauth2.NewClient(ctx, wrappedSource)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}

	return srv, nil
}

// GetMailboxes retrieves all mailboxes (labels) from Gmail
func (s *Service) GetMailboxes(ctx context.Context, accessToken, refreshToken string, onTokenRefresh TokenUpdateFunc) ([]*emaildomain.Mailbox, error) {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return nil, err
	}

	user := "me"
	labelsResp, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve labels: %v", err)
	}

	mailboxes := make([]*emaildomain.Mailbox, 0)

	// Map Gmail labels to our mailbox structure
	for _, label := range labelsResp.Labels {
		// Only include system labels and user labels
		if label.Type == "system" || label.Type == "user" {
			mailboxType := "user"
			if label.Type == "system" {
				mailboxType = strings.ToLower(label.Name)
			}
			
			mailbox := &emaildomain.Mailbox{
				ID:    label.Id,
				Name:  label.Name,
				Type:  mailboxType,
				Count: int(label.MessagesUnread),
			}
			mailboxes = append(mailboxes, mailbox)
		}
	}

	return mailboxes, nil
}

// GetEmails retrieves emails from a specific mailbox/label
func (s *Service) GetEmails(ctx context.Context, accessToken, refreshToken string, labelID string, limit, offset int, onTokenRefresh TokenUpdateFunc) ([]*emaildomain.Email, int, error) {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return nil, 0, err
	}

	user := "me"
	
	// Build query
	query := srv.Users.Messages.List(user)
	if labelID != "" && labelID != "ALL" {
		query = query.LabelIds(labelID)
	}
	query = query.MaxResults(int64(limit))

	messagesResp, err := query.Do()
	if err != nil {
		return nil, 0, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	emails := make([]*emaildomain.Email, 0)

	// Get full message details for each message
	for _, msg := range messagesResp.Messages {
		fullMsg, err := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()
		if err != nil {
			continue // Skip messages we can't fetch
		}

		email := convertGmailMessageToEmail(fullMsg)
		emails = append(emails, email)
	}

	return emails, int(messagesResp.ResultSizeEstimate), nil
}

// GetEmailByID retrieves a specific email by ID
func (s *Service) GetEmailByID(ctx context.Context, accessToken, refreshToken, emailID string, onTokenRefresh TokenUpdateFunc) (*emaildomain.Email, error) {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return nil, err
	}

	user := "me"
	msg, err := srv.Users.Messages.Get(user, emailID).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve message: %v", err)
	}

	return convertGmailMessageToEmail(msg), nil
}

// MarkAsRead marks an email as read
func (s *Service) MarkAsRead(ctx context.Context, accessToken, refreshToken, emailID string, onTokenRefresh TokenUpdateFunc) error {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return err
	}

	user := "me"
	modifyReq := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}

	_, err = srv.Users.Messages.Modify(user, emailID, modifyReq).Do()
	if err != nil {
		return fmt.Errorf("unable to mark message as read: %v", err)
	}

	return nil
}

// ToggleStar toggles the star status of an email
func (s *Service) ToggleStar(ctx context.Context, accessToken, refreshToken, emailID string, onTokenRefresh TokenUpdateFunc) error {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return err
	}

	user := "me"
	
	// Get current message to check star status
	msg, err := srv.Users.Messages.Get(user, emailID).Format("minimal").Do()
	if err != nil {
		return fmt.Errorf("unable to get message: %v", err)
	}

	isStarred := false
	for _, labelID := range msg.LabelIds {
		if labelID == "STARRED" {
			isStarred = true
			break
		}
	}

	var modifyReq *gmail.ModifyMessageRequest
	if isStarred {
		modifyReq = &gmail.ModifyMessageRequest{
			RemoveLabelIds: []string{"STARRED"},
		}
	} else {
		modifyReq = &gmail.ModifyMessageRequest{
			AddLabelIds: []string{"STARRED"},
		}
	}

	_, err = srv.Users.Messages.Modify(user, emailID, modifyReq).Do()
	if err != nil {
		return fmt.Errorf("unable to toggle star: %v", err)
	}

	return nil
}

// SendEmail sends an email
func (s *Service) SendEmail(ctx context.Context, accessToken, refreshToken string, to, subject, body string, onTokenRefresh TokenUpdateFunc) error {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return err
	}

	user := "me"
	
	// Create email message
	emailMsg := []byte(
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" + body)

	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(emailMsg),
	}

	_, err = srv.Users.Messages.Send(user, msg).Do()
	if err != nil {
		return fmt.Errorf("unable to send message: %v", err)
	}

	return nil
}

// TrashEmail moves an email to trash
func (s *Service) TrashEmail(ctx context.Context, accessToken, refreshToken, emailID string, onTokenRefresh TokenUpdateFunc) error {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return err
	}

	user := "me"
	modifyReq := &gmail.ModifyMessageRequest{
		AddLabelIds: []string{"TRASH"},
	}

	_, err = srv.Users.Messages.Modify(user, emailID, modifyReq).Do()
	if err != nil {
		return fmt.Errorf("unable to trash message: %v", err)
	}

	return nil
}

// ArchiveEmail archives an email (removes INBOX label)
func (s *Service) ArchiveEmail(ctx context.Context, accessToken, refreshToken, emailID string, onTokenRefresh TokenUpdateFunc) error {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return err
	}

	user := "me"
	modifyReq := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"INBOX"},
	}

	_, err = srv.Users.Messages.Modify(user, emailID, modifyReq).Do()
	if err != nil {
		return fmt.Errorf("unable to archive message: %v", err)
	}

	return nil
}

// Watch sets up push notifications for the user's mailbox
func (s *Service) Watch(ctx context.Context, accessToken, refreshToken string, topicName string, onTokenRefresh TokenUpdateFunc) error {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return err
	}

	// Try to stop any existing watch first to avoid "Only one user push notification client allowed" error
	// We ignore the error here because if there's no watch, it might fail, or if it succeeds, great.
	// But strictly speaking, we just want to ensure we clear the state if possible.
	log.Printf("Stopping existing watch for user...")
	_ = srv.Users.Stop("me").Do()

	req := &gmail.WatchRequest{
		TopicName: topicName,
		LabelIds:  []string{"INBOX"},
	}

	log.Printf("Starting watch for user on topic: %s", topicName)
	resp, err := srv.Users.Watch("me", req).Do()
	if err != nil {
		log.Printf("Gmail Watch API error: %v", err)
		return fmt.Errorf("unable to watch mailbox: %v", err)
	}
	log.Printf("Watch started successfully. Expiration: %d, HistoryId: %d", resp.Expiration, resp.HistoryId)

	return nil
}

// Stop stops push notifications for the user's mailbox
func (s *Service) Stop(ctx context.Context, accessToken, refreshToken string, onTokenRefresh TokenUpdateFunc) error {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return err
	}

	err = srv.Users.Stop("me").Do()
	if err != nil {
		return fmt.Errorf("unable to stop mailbox watch: %v", err)
	}

	return nil
}

// Helper functions

func convertGmailMessageToEmail(msg *gmail.Message) *emaildomain.Email {
	from := getHeader(msg.Payload.Headers, "From")
	fromName := from
	// Extract name from "Name <email@example.com>" format
	if idx := strings.Index(from, "<"); idx > 0 {
		fromName = strings.TrimSpace(from[:idx])
	}
	
	// Convert To header to array
	toHeader := getHeader(msg.Payload.Headers, "To")
	toArray := []string{}
	if toHeader != "" {
		toArray = []string{toHeader}
	}
	
	body := getEmailBody(msg.Payload)
	preview := body
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	
	email := &emaildomain.Email{
		ID:         msg.Id,
		Subject:    getHeader(msg.Payload.Headers, "Subject"),
		From:       from,
		FromName:   fromName,
		To:         toArray,
		Preview:    preview,
		Body:       body,
		IsHTML:     true, // Gmail usually returns HTML
		ReceivedAt: time.Unix(msg.InternalDate/1000, 0),
		IsRead:     !hasLabel(msg.LabelIds, "UNREAD"),
		IsStarred:  hasLabel(msg.LabelIds, "STARRED"),
		MailboxID:  getMailboxID(msg.LabelIds),
	}

	return email
}

func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, header := range headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}

func getEmailBody(payload *gmail.MessagePart) string {
	if payload.Body != nil && payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	// Check parts for body
	for _, part := range payload.Parts {
		if part.MimeType == "text/html" || part.MimeType == "text/plain" {
			if part.Body != nil && part.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					return string(data)
				}
			}
		}
		
		// Recursively check nested parts
		if len(part.Parts) > 0 {
			body := getEmailBody(part)
			if body != "" {
				return body
			}
		}
	}

	return ""
}

func hasLabel(labels []string, labelID string) bool {
	for _, label := range labels {
		if label == labelID {
			return true
		}
	}
	return false
}

func getMailboxID(labels []string) string {
	// Priority order for mailbox labels
	priority := []string{"INBOX", "SENT", "DRAFT", "SPAM", "TRASH"}
	
	for _, p := range priority {
		if hasLabel(labels, p) {
			return p
		}
	}
	
	// Return first label if no priority match
	if len(labels) > 0 {
		return labels[0]
	}
	
	return "INBOX"
}

func getIconForLabel(name string) string {
	iconMap := map[string]string{
		"INBOX":     "inbox",
		"SENT":      "send",
		"DRAFT":     "file-text",
		"STARRED":   "star",
		"SPAM":      "alert-circle",
		"TRASH":     "trash-2",
		"IMPORTANT": "bookmark",
		"CATEGORY_PERSONAL": "user",
		"CATEGORY_SOCIAL":   "users",
		"CATEGORY_PROMOTIONS": "tag",
		"CATEGORY_UPDATES":  "bell",
		"CATEGORY_FORUMS":   "message-square",
	}

	upperName := strings.ToUpper(name)
	if icon, ok := iconMap[upperName]; ok {
		return icon
	}

	return "folder"
}

// ValidateToken validates the access token by making a simple API call
func (s *Service) ValidateToken(ctx context.Context, accessToken, refreshToken string, onTokenRefresh TokenUpdateFunc) error {
	srv, err := s.GetGmailService(ctx, accessToken, refreshToken, onTokenRefresh)
	if err != nil {
		return err
	}

	_, err = srv.Users.GetProfile("me").Do()
	if err != nil {
		return errors.New("invalid or expired access token")
	}

	return nil
}
