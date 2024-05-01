package postgres

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"github.com/raystack/frontier/core/webhook"
	"github.com/raystack/frontier/pkg/crypt"
	"github.com/raystack/frontier/pkg/metadata"
)

type WebhookHeaders struct {
	KVs map[string]string `json:"kvs"`
}

func (s *WebhookHeaders) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, s)
	case string:
		return json.Unmarshal([]byte(src), s)
	case nil:
		return nil
	}
	return fmt.Errorf("cannot convert %T to JsonB", src)
}

func (s WebhookHeaders) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func toDBWebHookHeaders(headers map[string]string) WebhookHeaders {
	return WebhookHeaders{KVs: headers}
}

type WebhookSecret struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type WebhookSecrets struct {
	Items []WebhookSecret `json:"items"`
}

func toDBWebHookSecrets(secrets []webhook.Secret, encryptionKey []byte) (string, error) {
	var items []WebhookSecret
	for _, secret := range secrets {
		items = append(items, WebhookSecret{
			ID:    secret.ID,
			Value: secret.Value,
		})
	}
	dbSecretRawBytes, err := json.Marshal(WebhookSecrets{Items: items})
	if err != nil {
		return "", err
	}
	dbSecretEncryptedBytes, err := crypt.Encrypt(dbSecretRawBytes, encryptionKey)
	if err != nil {
		return "", err
	}

	// base64 encode the encrypted secret
	return base64.RawStdEncoding.EncodeToString(dbSecretEncryptedBytes), nil
}

func fromDBWebHookSecrets(secrets string, encryptionKey []byte) ([]webhook.Secret, error) {
	encryptedBytes, err := base64.RawStdEncoding.DecodeString(secrets)
	if err != nil {
		return nil, err
	}
	decryptedSecrets, err := crypt.Decrypt(encryptedBytes, encryptionKey)
	if err != nil {
		return nil, err
	}
	var dbSecrets WebhookSecrets
	if err := json.Unmarshal(decryptedSecrets, &dbSecrets); err != nil {
		return nil, err
	}
	var secretsList []webhook.Secret
	for _, secret := range dbSecrets.Items {
		secretsList = append(secretsList, webhook.Secret{
			ID:    secret.ID,
			Value: secret.Value,
		})
	}
	return secretsList, nil
}

type WebhookEndpoint struct {
	ID               string         `db:"id"`
	Description      *string        `db:"description"`
	SubscribedEvents pq.StringArray `db:"subscribed_events"`
	Headers          WebhookHeaders `db:"headers"`
	Url              string         `db:"url"`
	Secrets          string         `db:"secrets"`

	State    string             `db:"state"`
	Metadata types.NullJSONText `db:"metadata"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (i WebhookEndpoint) transform(encryptionKey []byte) (webhook.Endpoint, error) {
	var unmarshalledMetadata metadata.Metadata
	if i.Metadata.Valid {
		if err := i.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return webhook.Endpoint{}, err
		}
	}
	var description string
	if i.Description != nil {
		description = *i.Description
	}
	secrets, err := fromDBWebHookSecrets(i.Secrets, encryptionKey)
	if err != nil {
		return webhook.Endpoint{}, err
	}
	return webhook.Endpoint{
		ID:               i.ID,
		Description:      description,
		SubscribedEvents: i.SubscribedEvents,
		Secrets:          secrets,
		URL:              i.Url,
		Headers:          i.Headers.KVs,

		State:    webhook.State(i.State),
		Metadata: unmarshalledMetadata,

		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}, nil
}
