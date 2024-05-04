package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/raystack/frontier/core/webhook"
	"github.com/raystack/frontier/pkg/db"
)

type WebhookEndpointRepository struct {
	dbc           *db.Client
	encryptionKey []byte
}

func NewWebhookEndpointRepository(dbc *db.Client, encryptionKey []byte) *WebhookEndpointRepository {
	return &WebhookEndpointRepository{
		dbc:           dbc,
		encryptionKey: encryptionKey,
	}
}

func (r WebhookEndpointRepository) Create(ctx context.Context, toCreate webhook.Endpoint) (webhook.Endpoint, error) {
	if toCreate.ID == "" {
		toCreate.ID = uuid.New().String()
	}
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return webhook.Endpoint{}, err
	}
	secretString, err := toDBWebHookSecrets(toCreate.Secrets, r.encryptionKey)
	if err != nil {
		return webhook.Endpoint{}, fmt.Errorf("failed to encrypt webhook secrets: %w", err)
	}
	if toCreate.State == "" {
		toCreate.State = webhook.Enabled
	}

	query, params, err := dialect.Insert(TABLE_WEBHOOK_ENDPOINTS).Rows(
		goqu.Record{
			"id":                toCreate.ID,
			"description":       toCreate.Description,
			"subscribed_events": pq.StringArray(toCreate.SubscribedEvents),
			"secrets":           secretString,
			"headers":           toDBWebHookHeaders(toCreate.Headers),
			"url":               toCreate.URL,
			"state":             toCreate.State,
			"metadata":          marshaledMetadata,
			"created_at":        goqu.L("now()"),
			"updated_at":        goqu.L("now()"),
		}).Returning(&WebhookEndpoint{}).ToSQL()
	if err != nil {
		return webhook.Endpoint{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var endpointModel WebhookEndpoint
	if err = r.dbc.WithTimeout(ctx, TABLE_WEBHOOK_ENDPOINTS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&endpointModel)
	}); err != nil {
		return webhook.Endpoint{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return endpointModel.transform(r.encryptionKey)
}

func (r WebhookEndpointRepository) GetByID(ctx context.Context, id string) (webhook.Endpoint, error) {
	stmt := dialect.Select().From(TABLE_WEBHOOK_ENDPOINTS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return webhook.Endpoint{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var endpointModel WebhookEndpoint
	if err = r.dbc.WithTimeout(ctx, TABLE_WEBHOOK_ENDPOINTS, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&endpointModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return webhook.Endpoint{}, webhook.ErrNotFound
		}
		return webhook.Endpoint{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return endpointModel.transform(r.encryptionKey)
}

func (r WebhookEndpointRepository) List(ctx context.Context, flt webhook.EndpointFilter) ([]webhook.Endpoint, error) {
	stmt := dialect.Select().From(TABLE_WEBHOOK_ENDPOINTS).Order(goqu.I("created_at").Desc())
	if flt.State != "" {
		stmt = stmt.Where(goqu.Ex{
			"state": flt.State,
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var endpointModels []WebhookEndpoint
	if err = r.dbc.WithTimeout(ctx, TABLE_WEBHOOK_ENDPOINTS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &endpointModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	endpoints := make([]webhook.Endpoint, 0, len(endpointModels))
	for _, endpointModel := range endpointModels {
		endpoint, err := endpointModel.transform(r.encryptionKey)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (r WebhookEndpointRepository) UpdateByID(ctx context.Context, toUpdate webhook.Endpoint) (webhook.Endpoint, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return webhook.Endpoint{}, webhook.ErrInvalidDetail
	}

	updateRecord := goqu.Record{
		"description":       toUpdate.Description,
		"subscribed_events": pq.StringArray(toUpdate.SubscribedEvents),
		"url":               toUpdate.URL,
		"headers":           toDBWebHookHeaders(toUpdate.Headers),
		"updated_at":        goqu.L("now()"),
	}
	if toUpdate.Metadata != nil {
		marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
		if err != nil {
			return webhook.Endpoint{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		updateRecord["metadata"] = marshaledMetadata
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State
	}

	query, params, err := dialect.Update(TABLE_WEBHOOK_ENDPOINTS).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&WebhookEndpoint{}).ToSQL()
	if err != nil {
		return webhook.Endpoint{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var endpointModel WebhookEndpoint
	if err = r.dbc.WithTimeout(ctx, TABLE_WEBHOOK_ENDPOINTS, "UpdateByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&endpointModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return webhook.Endpoint{}, webhook.ErrNotFound
		default:
			return webhook.Endpoint{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return endpointModel.transform(r.encryptionKey)
}

func (r WebhookEndpointRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_WEBHOOK_ENDPOINTS).Where(goqu.Ex{
		"id": id,
	}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_WEBHOOK_ENDPOINTS, "Delete", func(ctx context.Context) error {
		_, err := r.dbc.ExecContext(ctx, query, params...)
		return err
	}); err != nil {
		return fmt.Errorf("%s: %w", txnErr, err)
	}
	return nil
}
