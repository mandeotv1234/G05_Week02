package usecase

import emaildomain "ga03-backend/internal/email/domain"

// EmailUsecase defines the interface for email use cases
type EmailUsecase interface {
	GetAllMailboxes(userID string) ([]*emaildomain.Mailbox, error)
	GetMailboxByID(id string) (*emaildomain.Mailbox, error)
	GetEmailsByMailbox(userID, mailboxID string, limit, offset int) ([]*emaildomain.Email, int, error)
	GetEmailByID(userID, id string) (*emaildomain.Email, error)
	MarkEmailAsRead(userID, id string) error
	ToggleStar(userID, id string) error
	SendEmail(userID, to, subject, body string) error
	TrashEmail(userID, id string) error
	ArchiveEmail(userID, id string) error
	WatchMailbox(userID string) error
}

