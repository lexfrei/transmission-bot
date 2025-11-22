// Package transmission provides a client wrapper for the Transmission RPC API.
package transmission

import (
	"context"
	"errors"
	"fmt"
	"time"

	gotransmission "github.com/lexfrei/go-transmission/api/transmission"

	"github.com/lexfrei/transmission-bot/internal/config"
)

// DefaultTimeout is the default timeout for Transmission RPC requests.
const DefaultTimeout = 30 * time.Second

// ErrUnexpectedResponse is returned when the Transmission API returns an unexpected response.
var ErrUnexpectedResponse = errors.New("unexpected response: no torrent added or duplicate")

// Client wraps the Transmission RPC client.
type Client struct {
	transmission gotransmission.Client
}

// Torrent represents a torrent in Transmission.
type Torrent struct {
	ID          int64
	Name        string
	Status      string
	PercentDone float64
	TotalSize   int64
}

// NewClient creates a new Transmission client with the given configuration.
func NewClient(cfg config.TransmissionConfig) (*Client, error) {
	opts := []gotransmission.Option{
		gotransmission.WithTimeout(DefaultTimeout),
	}

	if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts, gotransmission.WithAuth(cfg.Username, cfg.Password))
	}

	transmission, err := gotransmission.New(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating transmission client: %w", err)
	}

	return &Client{transmission: transmission}, nil
}

// Close releases resources associated with the client.
func (c *Client) Close() error {
	closeErr := c.transmission.Close()
	if closeErr != nil {
		return fmt.Errorf("closing transmission client: %w", closeErr)
	}

	return nil
}

// AddTorrentByMagnet adds a torrent using a magnet link.
func (c *Client) AddTorrentByMagnet(ctx context.Context, magnet string) (*Torrent, error) {
	result, err := c.transmission.TorrentAdd(ctx, &gotransmission.TorrentAddArgs{
		Filename: &magnet,
	})
	if err != nil {
		return nil, fmt.Errorf("adding torrent: %w", err)
	}

	if result.TorrentAdded != nil {
		return &Torrent{
			ID:   result.TorrentAdded.ID,
			Name: result.TorrentAdded.Name,
		}, nil
	}

	if result.TorrentDuplicate != nil {
		return &Torrent{
			ID:   result.TorrentDuplicate.ID,
			Name: result.TorrentDuplicate.Name,
		}, nil
	}

	return nil, ErrUnexpectedResponse
}

// AddTorrentByFile adds a torrent using base64-encoded torrent file data.
func (c *Client) AddTorrentByFile(ctx context.Context, base64Data string) (*Torrent, error) {
	result, err := c.transmission.TorrentAdd(ctx, &gotransmission.TorrentAddArgs{
		Metainfo: &base64Data,
	})
	if err != nil {
		return nil, fmt.Errorf("adding torrent: %w", err)
	}

	if result.TorrentAdded != nil {
		return &Torrent{
			ID:   result.TorrentAdded.ID,
			Name: result.TorrentAdded.Name,
		}, nil
	}

	if result.TorrentDuplicate != nil {
		return &Torrent{
			ID:   result.TorrentDuplicate.ID,
			Name: result.TorrentDuplicate.Name,
		}, nil
	}

	return nil, ErrUnexpectedResponse
}

// ListTorrents returns a list of all torrents.
func (c *Client) ListTorrents(ctx context.Context) ([]Torrent, error) {
	fields := []string{"id", "name", "status", "percentDone", "totalSize"}

	result, err := c.transmission.TorrentGet(ctx, fields, nil)
	if err != nil {
		return nil, fmt.Errorf("getting torrents: %w", err)
	}

	torrents := make([]Torrent, 0, len(result.Torrents))
	for _, torrent := range result.Torrents {
		torrents = append(torrents, Torrent{
			ID:          *torrent.ID,
			Name:        *torrent.Name,
			Status:      torrent.Status.String(),
			PercentDone: *torrent.PercentDone,
			TotalSize:   *torrent.TotalSize,
		})
	}

	return torrents, nil
}

// ErrTorrentNotFound is returned when the requested torrent does not exist.
var ErrTorrentNotFound = errors.New("torrent not found")

// GetTorrent returns a torrent by ID.
func (c *Client) GetTorrent(ctx context.Context, torrentID int64) (*Torrent, error) {
	fields := []string{"id", "name", "status", "percentDone", "totalSize"}

	result, err := c.transmission.TorrentGet(ctx, fields, []int64{torrentID})
	if err != nil {
		return nil, fmt.Errorf("getting torrent: %w", err)
	}

	if len(result.Torrents) == 0 {
		return nil, ErrTorrentNotFound
	}

	torrent := result.Torrents[0]

	return &Torrent{
		ID:          *torrent.ID,
		Name:        *torrent.Name,
		Status:      torrent.Status.String(),
		PercentDone: *torrent.PercentDone,
		TotalSize:   *torrent.TotalSize,
	}, nil
}

// RemoveTorrent removes a torrent by ID, optionally deleting local data.
func (c *Client) RemoveTorrent(ctx context.Context, torrentID int64, deleteData bool) error {
	err := c.transmission.TorrentRemove(ctx, []int64{torrentID}, deleteData)
	if err != nil {
		return fmt.Errorf("removing torrent: %w", err)
	}

	return nil
}
