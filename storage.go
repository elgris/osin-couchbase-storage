package storage

import (
	"errors"

	"github.com/RangelReale/osin"
	"github.com/couchbase/gocb"
)

const (
	accessTokenKeyPrefix  = "a_"
	refreshTokenKeyPrefix = "r_"
)

type Config struct {
	ConnectionString string
	BucketName       string
	BucketPassword   string
}

func (c Config) Validate() error {
	if len(c.ConnectionString) == 0 {
		return errors.New("empty connection string provided")
	}

	if len(c.BucketName) == 0 {
		return errors.New("empty bucket name provided")
	}

	return nil
}

type Storage struct {
	bucket *gocb.Bucket
	// TODO: logger
}

func NewStorage(config Config) (*Storage, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	conn, err := gocb.Connect(config.ConnectionString)
	if err != nil {
		return nil, err
	}

	bucket, err := conn.OpenBucket(config.BucketName, config.BucketPassword)
	if err != nil {
		return nil, err
	}

	s := Storage{
		bucket: bucket,
	}

	return &s, nil
}

// Clone implements osin.Storage interface, but in fact does not
// clone the storage
func (s *Storage) Clone() osin.Storage { return s }

// Close closes connections and cleans up resources
func (s *Storage) Close() {
	// s.bucket.Close()
}

// SetClient saves client record to the storage. Client record must provide
// Id in client.Id
func (s *Storage) SetClient(client osin.Client) error {
	_, err := s.bucket.Upsert(client.GetId(), client, 0)

	return err
}

// GetClient loads the client by id (client_id)
func (s *Storage) GetClient(id string) (osin.Client, error) {
	client := &osin.DefaultClient{}
	_, err := s.bucket.Get(id, client)

	return client, err
}

// SaveAuthorize saves authorize data.
func (s *Storage) SaveAuthorize(data *osin.AuthorizeData) error {
	_, err := s.bucket.Upsert(data.Code, data, uint32(data.ExpiresIn))

	return err
}

// LoadAuthorize looks up AuthorizeData by a code.
// Client information MUST be loaded together.
// Optionally can return error if expired.
func (s *Storage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	data := &osin.AuthorizeData{}
	data.Client = &osin.DefaultClient{}

	_, err := s.bucket.Get(code, data)

	return data, err
}

// RemoveAuthorize revokes or deletes the authorization code.
func (s *Storage) RemoveAuthorize(code string) error {
	_, err := s.bucket.Remove(code, 0)

	return err
}

// SaveAccess writes AccessData.
// If RefreshToken is not blank, it must save in a way that can be loaded using LoadRefresh.
func (s *Storage) SaveAccess(data *osin.AccessData) error {
	// save access token
	accessKey := accessTokenKeyPrefix + data.AccessToken
	if _, err := s.bucket.Upsert(accessKey, data, uint32(data.ExpiresIn)); err != nil {
		return err
	}

	// save refresh token if present
	if len(data.RefreshToken) > 0 && data.AccessData != nil {
		refreshKey := refreshTokenKeyPrefix + data.RefreshToken

		if _, err := s.bucket.Upsert(refreshKey, data.AccessData, uint32(data.ExpiresIn)); err != nil {
			return err
		}
	}

	return nil
}

// LoadAccess retrieves access data by token.
func (s *Storage) LoadAccess(token string) (*osin.AccessData, error) {
	key := accessTokenKeyPrefix + token
	data := &osin.AccessData{
		Client: &osin.DefaultClient{},
		AuthorizeData: &osin.AuthorizeData{
			Client: &osin.DefaultClient{},
		},
		AccessData: &osin.AccessData{
			Client: &osin.DefaultClient{},
			AuthorizeData: &osin.AuthorizeData{
				Client: &osin.DefaultClient{},
			},
		},
	}

	_, err := s.bucket.Get(key, data)
	return data, err
}

// RemoveAccess revokes or deletes an AccessData.
func (s *Storage) RemoveAccess(token string) error {
	key := accessTokenKeyPrefix + token
	_, err := s.bucket.Remove(key, 0)

	return err
}

// LoadRefresh retrieves refresh AccessData. Client information MUST be loaded together.
// AuthorizeData and AccessData DON'T NEED to be loaded if not easily available.
// Optionally can return error if expired.
func (s *Storage) LoadRefresh(token string) (*osin.AccessData, error) {
	key := refreshTokenKeyPrefix + token
	data := &osin.AccessData{
		Client: &osin.DefaultClient{},
		AuthorizeData: &osin.AuthorizeData{
			Client: &osin.DefaultClient{},
		},
		AccessData: &osin.AccessData{
			Client: &osin.DefaultClient{},
			AuthorizeData: &osin.AuthorizeData{
				Client: &osin.DefaultClient{},
			},
		},
	}

	_, err := s.bucket.Get(key, data)
	return data, err
}

// RemoveRefresh revokes or deletes refresh AccessData.
func (s *Storage) RemoveRefresh(token string) error {
	key := refreshTokenKeyPrefix + token
	_, err := s.bucket.Remove(key, 0)

	return err
}
