package storage

import (
	"flag"
	"testing"
	"time"

	"github.com/RangelReale/osin"
	"github.com/couchbase/gocb/gocbcore"
	"github.com/stretchr/testify/assert"
)

var (
	connString     = flag.String("couchbase", "couchbase://docker", "Connection string for Couchbase DB")
	bucketName     = flag.String("bucket", "default", "Name of the bucket for test")
	bucketPassword = flag.String("password", "", "Password for the bucket")

	storage *Storage
)

func TestIncorrectConfig(t *testing.T) {
	data := []struct {
		Config   Config
		Expected string
	}{
		{Config{ConnectionString: "", BucketName: "something"}, "empty connection string provided"},
		{Config{ConnectionString: "something", BucketName: ""}, "empty bucket name provided"},
	}

	for _, item := range data {
		_, err := NewStorage(item.Config)
		assert.Error(t, err)
		assert.Equal(t, item.Expected, err.Error())
	}
}

func TestGetClient(t *testing.T) {
	clientData := &osin.DefaultClient{
		Id:          "client_id",
		Secret:      "client_secret",
		RedirectUri: "http://some.redirect.here",
		UserData:    "foobar",
	}

	storage := getStorage()
	err := storage.SetClient(clientData)
	assert.NoError(t, err)

	actualClient, err := storage.GetClient(clientData.Id)
	assert.NoError(t, err)
	assert.Equal(t, clientData, actualClient)
}

func TestSaveLoadRemoveAuthorize(t *testing.T) {
	createdAd, err := time.Parse(time.RFC3339, "2015-01-01T12:12:12Z")
	assert.NoError(t, err)

	data := &osin.AuthorizeData{
		Client: &osin.DefaultClient{
			Id:          "client_id",
			Secret:      "client_secret",
			RedirectUri: "http://some.redirect.here",
			UserData:    "foobar",
		},
		Code:        "foo",
		ExpiresIn:   1000,
		Scope:       "scope.foo.read",
		RedirectUri: "http://redirect.me",
		State:       "state.foo",
		CreatedAt:   createdAd,
		UserData:    "foodata",
	}

	storage := getStorage()
	// save
	err = storage.SaveAuthorize(data)
	assert.NoError(t, err)

	// load
	actual, err := storage.LoadAuthorize(data.Code)
	assert.NoError(t, err)
	assert.Equal(t, data, actual)

	// remove
	err = storage.RemoveAuthorize(data.Code)
	assert.NoError(t, err)
	_, err = storage.LoadAuthorize(data.Code)
	assert.Equal(t, gocbcore.ErrKeyNotFound, err)
}

func TestSaveLoadRemoveAccessRefresh(t *testing.T) {
	createdAd, err := time.Parse(time.RFC3339, "2015-01-01T12:12:12Z")
	assert.NoError(t, err)

	refreshData := &osin.AccessData{
		Client: &osin.DefaultClient{
			Id:          "client_id",
			Secret:      "client_secret",
			RedirectUri: "http://some.redirect.here",
			UserData:    "foobar",
		},
		AuthorizeData: &osin.AuthorizeData{
			Client: &osin.DefaultClient{
				Id:          "client_id",
				Secret:      "client_secret",
				RedirectUri: "http://some.redirect.here",
				UserData:    "foobar",
			},
			Code:        "refreshfoo",
			ExpiresIn:   1234,
			Scope:       "scope.foo.read",
			RedirectUri: "http://redirect.me",
			State:       "state.foo",
			CreatedAt:   createdAd,
			UserData:    "refreshfoodata",
		},
		AccessToken: "refreshfootookeeen",
		ExpiresIn:   1000,
		Scope:       "scope.foo.read",
		RedirectUri: "http://redirect.me",
		CreatedAt:   createdAd,
		UserData:    "refreshfoodata",
	}

	accessData := &osin.AccessData{
		Client: &osin.DefaultClient{
			Id:          "client_id",
			Secret:      "client_secret",
			RedirectUri: "http://some.redirect.here",
			UserData:    "foobar",
		},
		AuthorizeData: &osin.AuthorizeData{
			Client: &osin.DefaultClient{
				Id:          "client_id",
				Secret:      "client_secret",
				RedirectUri: "http://some.redirect.here",
				UserData:    "foobar",
			},
			Code:        "foo",
			ExpiresIn:   1000,
			Scope:       "scope.foo.read",
			RedirectUri: "http://redirect.me",
			State:       "state.foo",
			CreatedAt:   createdAd,
			UserData:    "foodata",
		},
		AccessData:   refreshData,
		RefreshToken: "refreshfootookeeen",
		AccessToken:  "footookeeen",
		ExpiresIn:    4321,
		Scope:        "scope.foo.read",
		RedirectUri:  "http://redirect.me",
		CreatedAt:    createdAd,
		UserData:     "accessfoodata",
	}

	storage := getStorage()
	// save
	err = storage.SaveAccess(accessData)
	assert.NoError(t, err)

	// load access token
	actual, err := storage.LoadAccess(accessData.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessData, actual)

	// load refresh token
	actual, err = storage.LoadRefresh(accessData.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, refreshData, actual)

	// remove refresh token
	err = storage.RemoveRefresh(accessData.RefreshToken)
	assert.NoError(t, err)
	_, err = storage.LoadRefresh(accessData.RefreshToken)
	assert.Equal(t, gocbcore.ErrKeyNotFound, err)

	// remove access token
	err = storage.RemoveAccess(accessData.AccessToken)
	assert.NoError(t, err)
	_, err = storage.LoadAccess(accessData.AccessToken)
	assert.Equal(t, gocbcore.ErrKeyNotFound, err)
}

func getStorage() *Storage {
	if storage != nil {
		return storage
	}

	var err error
	storage, err = NewStorage(Config{
		ConnectionString: *connString,
		BucketName:       *bucketName,
		BucketPassword:   *bucketPassword,
	})

	if err != nil {
		panic("could not communicate to Couchbase: " + err.Error())
	}

	return storage
}
