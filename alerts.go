package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// FingguAlertChannel is implemented by every notification backend
// (console, Slack, Discord, ...). Adding a new channel means implementing
// this one method — see FingguSlackChannel for an example.
type FingguAlertChannel interface {
	Send(message string) error
	Name() string
}

// FingguAlertManager fans a bottleneck result out to every configured
// channel, with a per-reason-set cooldown so a sustained issue doesn't
// spam the same alert every poll interval.
type FingguAlertManager struct {
	channels []FingguAlertChannel
	cooldown time.Duration
	mu       sync.Mutex
	lastSent map[string]time.Time
}

// FingguNewAlertManager builds the manager and registers whichever channels
// have credentials configured. Console output is always enabled.
func FingguNewAlertManager(cfg *FingguConfig) *FingguAlertManager {
	channels := []FingguAlertChannel{&FingguConsoleChannel{}}

	if cfg.SlackWebhookURL != "" {
		channels = append(channels, &FingguSlackChannel{WebhookURL: cfg.SlackWebhookURL})
	}
	if cfg.DiscordWebhookURL != "" {
		channels = append(channels, &FingguDiscordChannel{WebhookURL: cfg.DiscordWebhookURL})
	}

	return &FingguAlertManager{
		channels: channels,
		cooldown: cfg.AlertCooldown,
		lastSent: make(map[string]time.Time),
	}
}

// FingguDispatch sends the alert to every registered channel, respecting
// the cooldown window keyed by the combined reason text.
func (manager *FingguAlertManager) FingguDispatch(result FingguBottleneckResult) {
	key := strings.Join(result.Reasons, "|")

	manager.mu.Lock()
	if last, ok := manager.lastSent[key]; ok && time.Since(last) < manager.cooldown {
		manager.mu.Unlock()
		return
	}
	manager.lastSent[key] = time.Now()
	manager.mu.Unlock()

	message := fmt.Sprintf("⚠️ Finggu Performance Alert\nTime: %s\n%s",
		result.Metric.Timestamp.Format(time.RFC3339), strings.Join(result.Reasons, "\n"))

	for _, channel := range manager.channels {
		if err := channel.Send(message); err != nil {
			log.Printf("finggu: alert channel %s failed: %v", channel.Name(), err)
		}
	}
}

// FingguConsoleChannel is the always-on fallback channel; zero setup required.
type FingguConsoleChannel struct{}

func (c *FingguConsoleChannel) Send(message string) error {
	log.Println(message)
	return nil
}

func (c *FingguConsoleChannel) Name() string { return "console" }

// FingguSlackChannel posts to an incoming Slack webhook URL.
type FingguSlackChannel struct {
	WebhookURL string
}

func (c *FingguSlackChannel) Send(message string) error {
	return fingguFn_PostWebhookJSON(c.WebhookURL, map[string]string{"text": message})
}

func (c *FingguSlackChannel) Name() string { return "slack" }

// FingguDiscordChannel posts to a Discord incoming webhook URL.
type FingguDiscordChannel struct {
	WebhookURL string
}

func (c *FingguDiscordChannel) Send(message string) error {
	return fingguFn_PostWebhookJSON(c.WebhookURL, map[string]string{"content": message})
}

func (c *FingguDiscordChannel) Name() string { return "discord" }

func fingguFn_PostWebhookJSON(url string, payload map[string]string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
