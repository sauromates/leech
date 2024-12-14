package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	btmessage "gihub.com/sauromates/leech/internal"
	"gihub.com/sauromates/leech/internal/utils"
)

const (
	// MaxBlockSize is the largest number of bytes a request can ask for
	MaxBlockSize = 16000
	// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
	MaxBacklog = 50
)

type Worker interface {
	Start(*Torrent, utils.Peer, chan *TaskItem, chan *TaskResult)
	GetBasePath() string
	download(*Client, *TaskItem) ([]byte, error)
}

type DownloadWorker struct {
	BasePath string
}

func (worker *DownloadWorker) GetBasePath() string {
	return worker.BasePath
}

func (worker *DownloadWorker) Start(torrent *Torrent, peer utils.Peer, queue chan *TaskItem, results chan *TaskResult) {
	client, err := NewClient(peer, torrent.InfoHash, torrent.PeerID)
	if err != nil {
		log.Printf("Could not handshake with %s (%s). Disconnecting\n", peer.IP, err)
		return
	}

	defer client.Conn.Close()

	log.Printf("Completed handshake with %s\n", peer.IP)
	log.Printf("Announcing interest to peer %s", peer.String())

	client.SendMessage(btmessage.Unchoke)
	client.SendMessage(btmessage.Interested)

	for task := range queue {
		// Put task back into queue, peer doesn't have this piece
		if !client.BitField.HasPiece(task.Index) {
			queue <- task
			continue
		}

		content, err := worker.download(client, task)
		if err != nil {
			queue <- task

			// log.Printf(
			// 	"Download failed, putting task %d back into queue (new size is %d). Reason: %s",
			// 	task.Index,
			// 	len(queue),
			// 	err,
			// )

			return
		}

		if err := verify(task, content); err != nil {
			log.Printf("piece #%d failed integrity check\n", task.Index)
			queue <- task

			continue
		}

		client.SendHave(task.Index)

		results <- &TaskResult{task.Index, content}
	}
}

func (worker *DownloadWorker) download(client *Client, task *TaskItem) ([]byte, error) {
	state := TaskProgress{
		Index:   task.Index,
		Client:  client,
		Content: make([]byte, task.Length),
	}

	client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer client.Conn.SetDeadline(time.Time{})

	for state.Downloaded < task.Length {
		if !state.Client.IsChoked {
			for state.Backlog < MaxBacklog && state.Requested < task.Length {
				blockSize := MaxBlockSize
				// Last block might be shorter than the typical block
				if task.Length-state.Requested < blockSize {
					blockSize = task.Length - state.Requested
				}

				if err := client.SendRequest(task.Index, state.Requested, blockSize); err != nil {
					return nil, fmt.Errorf("request failed %s", err)
				}

				state.Backlog++
				state.Requested += blockSize
			}
		}

		if err := state.ReadMessage(); err != nil {
			return nil, fmt.Errorf("read failed %s", err)
		}
	}

	return state.Content, nil
}

func verify(task *TaskItem, content []byte) error {
	hash := sha1.Sum(content)
	if !bytes.Equal(hash[:], task.Hash[:]) {
		return fmt.Errorf("index %d failed integrity check", task.Index)
	}

	return nil
}
