package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	authrepo "ga03-backend/internal/auth/repository"
	"ga03-backend/pkg/sse"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type GmailNotification struct {
	EmailAddress string `json:"emailAddress"`
	HistoryID    uint64 `json:"historyId"`
}

type Service struct {
	pubsubClient *pubsub.Client
	sseManager   *sse.Manager
	userRepo     authrepo.UserRepository
	projectID    string
	topicName    string
	subName      string
}

func NewService(projectID, topicName string, sseManager *sse.Manager, userRepo authrepo.UserRepository, credentialsFile string) (*Service, error) {
	ctx := context.Background()
	
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %v", err)
	}

	return &Service{
		pubsubClient: client,
		sseManager:   sseManager,
		userRepo:     userRepo,
		projectID:    projectID,
		topicName:    topicName,
		subName:      topicName + "-sub", // Convention: topic-sub
	}, nil
}

func (s *Service) Start(ctx context.Context) {
	// Ensure subscription exists
	sub := s.pubsubClient.Subscription(s.subName)
	exists, err := sub.Exists(ctx)
	if err != nil {
		log.Printf("Error checking subscription existence: %v", err)
		return
	}

	if !exists {
		topic := s.pubsubClient.Topic(s.topicName)
		sub, err = s.pubsubClient.CreateSubscription(ctx, s.subName, pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 10 * time.Second,
		})
		if err != nil {
			log.Printf("Failed to create subscription: %v", err)
			return
		}
		log.Printf("Created subscription: %s", s.subName)
	}

	log.Printf("Listening for messages on subscription: %s", s.subName)
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		s.handleMessage(ctx, msg)
		msg.Ack()
	})
	if err != nil {
		log.Printf("Error receiving messages: %v", err)
	}
}

func (s *Service) handleMessage(ctx context.Context, msg *pubsub.Message) {
	var notification GmailNotification
	if err := json.Unmarshal(msg.Data, &notification); err != nil {
		log.Printf("Failed to unmarshal notification: %v", err)
		return
	}

	log.Printf("Received notification for: %s", notification.EmailAddress)

	// Find user by email
	user, err := s.userRepo.FindByEmail(notification.EmailAddress)
	if err != nil {
		log.Printf("Error finding user by email %s: %v", notification.EmailAddress, err)
		return
	}
	if user == nil {
		log.Printf("User not found for email: %s", notification.EmailAddress)
		return
	}

	// Notify user via SSE
	s.sseManager.SendToUser(user.ID, "email_update", map[string]interface{}{
		"email":     notification.EmailAddress,
		"historyId": notification.HistoryID,
		"timestamp": time.Now(),
	})
}
