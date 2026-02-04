package handover

import (
	"context"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type NotificationService interface {
	SendDriverCheckinNotification(ctx context.Context, msg string) error
}

type waService struct {
	client  *whatsmeow.Client
	groupID string // JID Group WA (format: 123456789@g.us)
}

func NewWANotificationService(client *whatsmeow.Client, groupID string) NotificationService {
	return &waService{client: client, groupID: groupID}
}

func (w *waService) SendDriverCheckinNotification(ctx context.Context, msg string) error {
	groupJID, err := types.ParseJID(w.groupID)
	if err != nil {
		return err
	}

	_, err = w.client.SendMessage(ctx, groupJID, &waE2E.Message{
		Conversation: proto.String(msg),
	})
	return err
}
