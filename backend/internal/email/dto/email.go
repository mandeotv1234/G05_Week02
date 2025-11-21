package dto

import emaildomain "ga03-backend/internal/email/domain"

type MailboxesResponse struct {
	Mailboxes []*emaildomain.Mailbox `json:"mailboxes"`
}

type EmailsResponse struct {
	Emails []*emaildomain.Email `json:"emails"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
	Total  int                  `json:"total"`
}

type SendEmailRequest struct {
	To      string `json:"to" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

