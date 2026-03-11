// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/timlinux/cheetah/backend"
)

// Client communicates with the backend REST API
type Client struct {
	baseURL   string
	sessionID string
	client    *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WaitForServer waits for the server to be ready
func (c *Client) WaitForServer(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := c.client.Get(c.baseURL + "/api/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server not ready after %v", timeout)
}

// CreateSession creates a new session on the server
func (c *Client) CreateSession() error {
	resp, err := c.client.Post(c.baseURL+"/api/v1/sessions", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create session: %s", resp.Status)
	}

	var result struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	c.sessionID = result.SessionID
	return nil
}

// DeleteSession deletes the current session
func (c *Client) DeleteSession() error {
	if c.sessionID == "" {
		return nil
	}

	req, err := http.NewRequest("DELETE", c.baseURL+"/api/v1/sessions/"+c.sessionID, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// LoadDocument loads a document from a file path
func (c *Client) LoadDocument(path string) error {
	body := bytes.NewBufferString(fmt.Sprintf(`{"path":"%s"}`, path))
	resp, err := c.client.Post(
		c.baseURL+"/api/v1/sessions/"+c.sessionID+"/document/path",
		"application/json",
		body,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to load document: %s - %s", resp.Status, string(body))
	}

	return nil
}

// UploadDocument uploads a document file
func (c *Client) UploadDocument(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("document", filepath.Base(path))
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	writer.Close()

	resp, err := c.client.Post(
		c.baseURL+"/api/v1/sessions/"+c.sessionID+"/document",
		writer.FormDataContentType(),
		body,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload document: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

// GetDocumentInfo returns information about the loaded document
func (c *Client) GetDocumentInfo() (*backend.DocumentInfo, error) {
	resp, err := c.client.Get(c.baseURL + "/api/v1/sessions/" + c.sessionID + "/document/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get document info: %s", resp.Status)
	}

	var info backend.DocumentInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetState returns the current reading state
func (c *Client) GetState() (*backend.ReadingState, error) {
	resp, err := c.client.Get(c.baseURL + "/api/v1/sessions/" + c.sessionID + "/state")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get state: %s", resp.Status)
	}

	var state backend.ReadingState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}

	return &state, nil
}

// Play starts reading
func (c *Client) Play() error {
	resp, err := c.client.Post(c.baseURL+"/api/v1/sessions/"+c.sessionID+"/play", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Pause pauses reading
func (c *Client) Pause() error {
	resp, err := c.client.Post(c.baseURL+"/api/v1/sessions/"+c.sessionID+"/pause", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Toggle toggles play/pause
func (c *Client) Toggle() error {
	resp, err := c.client.Post(c.baseURL+"/api/v1/sessions/"+c.sessionID+"/toggle", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// SetWPM sets the reading speed
func (c *Client) SetWPM(wpm int) error {
	body := bytes.NewBufferString(fmt.Sprintf(`{"wpm":%d}`, wpm))
	resp, err := c.client.Post(
		c.baseURL+"/api/v1/sessions/"+c.sessionID+"/speed",
		"application/json",
		body,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// PrevParagraph moves to the previous paragraph
func (c *Client) PrevParagraph() error {
	resp, err := c.client.Post(c.baseURL+"/api/v1/sessions/"+c.sessionID+"/paragraph/prev", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// NextParagraph moves to the next paragraph
func (c *Client) NextParagraph() error {
	resp, err := c.client.Post(c.baseURL+"/api/v1/sessions/"+c.sessionID+"/paragraph/next", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// JumpToWord moves to a specific word
func (c *Client) JumpToWord(index int) error {
	resp, err := c.client.Post(
		fmt.Sprintf("%s/api/v1/sessions/%s/word/%d", c.baseURL, c.sessionID, index),
		"application/json",
		nil,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// SavePosition saves the current reading position
func (c *Client) SavePosition() error {
	resp, err := c.client.Post(c.baseURL+"/api/v1/sessions/"+c.sessionID+"/save", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// GetSavedSessions returns all saved sessions
func (c *Client) GetSavedSessions() ([]backend.SavedSession, error) {
	resp, err := c.client.Get(c.baseURL + "/api/v1/saved")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessions []backend.SavedSession
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// ReturnToStart returns to the beginning of the document
func (c *Client) ReturnToStart() error {
	return c.JumpToWord(0)
}

// ResumeSession resumes a saved session
func (c *Client) ResumeSession(hash string) error {
	resp, err := c.client.Post(
		c.baseURL+"/api/v1/saved/"+hash+"/resume",
		"application/json",
		nil,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to resume session: %s", resp.Status)
	}

	var result struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	c.sessionID = result.SessionID
	return nil
}
