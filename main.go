package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"log"
	"os"
	"sort"
	"time"
)

const delayTime = time.Millisecond * 100

func main() {
	// Get channel name
	if len(os.Args) < 2 {
		log.Fatalln("Please pass the channel ID as the argument")
	}
	channelUsername := os.Args[1]
	// Setup bot
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: "session",
		},
	})
	if err != nil {
		log.Fatalln("cannot create telegram client:", err)
	}
	// Get the phone number
	phoneNumber := os.Getenv("PHONE")
	if phoneNumber == "" {
		log.Fatalln("please provide your phone number as \"PHONE\" environment variable")
	}
	err = client.Run(context.Background(), func(ctx context.Context) error {
		err := client.Auth().IfNecessary(ctx, auth.NewFlow(
			SimpleAuth{PhoneNumber: phoneNumber},
			auth.SendCodeOptions{},
		))
		if err != nil {
			return err
		}
		api := client.API()
		// Get channel
		channel, err := getChannel(ctx, api, channelUsername)
		if err != nil {
			return fmt.Errorf("cannot get channel: %w", err)
		}
		// Get history
		history, err := getAllHistory(ctx, api, &tg.InputPeerChannel{
			ChannelID:  channel.ID,
			AccessHash: channel.AccessHash,
		})
		if err != nil {
			return err
		}
		sort.Sort(history)
		// Write to file I guess
		return processSortedHistory(history, channelUsername)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func getChannel(ctx context.Context, api *tg.Client, username string) (*tg.Channel, error) {
	data, err := api.ContactsResolveUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	// Check if it exists
	if len(data.Chats) == 0 {
		return nil, errors.New("channel not found")

	}
	c, ok := data.Chats[0].(*tg.Channel)
	if !ok {
		return nil, errors.New("chat is not channel")
	}
	// Create the channel
	return c, nil
}

func getAllHistory(ctx context.Context, api *tg.Client, channel tg.InputPeerClass) (Messages, error) {
	req := &tg.MessagesGetHistoryRequest{
		Peer:       channel,
		OffsetID:   0,
		OffsetDate: 0,
		AddOffset:  0,
		Limit:      0,
		MaxID:      0,
		MinID:      0,
		Hash:       0,
	}
	// Get the first messages
	historyClass, err := api.MessagesGetHistory(ctx, req)
	if err != nil {
		return nil, err
	}
	history := historyClass.(*tg.MessagesChannelMessages)
	// Create the result
	result := make([]Message, 0, history.Count)
	for len(history.Messages) > 0 {
		fmt.Printf("Progress %d out of %d\n", history.OffsetIDOffset, history.Count)
		// Append current messages
		for _, msg := range history.Messages {
			message, ok := msg.(*tg.Message)
			if !ok {
				continue
			}
			result = append(result, Message{
				Id:    message.ID,
				Views: message.Views,
			})
		}
		// Get messages before
		req.OffsetID = history.Messages[len(history.Messages)-1].GetID()
		// Request more
		for {
			time.Sleep(delayTime)
			historyClass, err = api.MessagesGetHistory(ctx, req)
			if err != nil {
				if flood, err := tgerr.FloodWait(ctx, err); err != nil {
					if flood || tgerr.Is(err, tg.ErrTimeout) {
						continue
					}
				}
				return nil, err
			}
			break
		}
		history = historyClass.(*tg.MessagesChannelMessages)
	}
	return result, nil
}

func processSortedHistory(messages Messages, channelUsername string) error {
	f, err := os.Create("views.txt")
	if err != nil {
		return fmt.Errorf("cannot open file for write: %w", err)
	}
	defer f.Close()
	// Write messages
	for _, msg := range messages {
		_, err = fmt.Fprintf(f, "https://t.me/%s/%d\n", channelUsername, msg.Id)
		if err != nil {
			return fmt.Errorf("cannot write to file: %w", err)
		}
	}
	return nil
}
