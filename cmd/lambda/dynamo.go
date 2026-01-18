package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/colby/snip/internal/model"
	"github.com/colby/snip/internal/repository"
)

// DynamoLinkRepository implements repository.LinkRepository using DynamoDB.
type DynamoLinkRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoLinkRepository creates a new DynamoDB-backed link repository.
func NewDynamoLinkRepository(tableName string) *DynamoLinkRepository {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	return &DynamoLinkRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}
}

// Create stores a new link in DynamoDB.
func (r *DynamoLinkRepository) Create(ctx context.Context, link *model.Link) error {
	item := map[string]types.AttributeValue{
		"short_code":   &types.AttributeValueMemberS{Value: link.ShortCode},
		"original_url": &types.AttributeValueMemberS{Value: link.OriginalURL},
		"created_at":   &types.AttributeValueMemberS{Value: link.CreatedAt.Format(time.RFC3339)},
		"click_count":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", link.ClickCount)},
	}

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(short_code)"),
	})

	if err != nil {
		// Check if it failed because the item already exists
		var condErr *types.ConditionalCheckFailedException
		if ok := errors.As(err, &condErr); ok {
			return repository.ErrAlreadyExists
		}
		return fmt.Errorf("dynamodb put item: %w", err)
	}

	return nil
}

// GetByShortCode retrieves a link by its short code.
func (r *DynamoLinkRepository) GetByShortCode(ctx context.Context, shortCode string) (*model.Link, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"short_code": &types.AttributeValueMemberS{Value: shortCode},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("dynamodb get item: %w", err)
	}

	if result.Item == nil {
		return nil, repository.ErrNotFound
	}

	link, err := itemToLink(result.Item)
	if err != nil {
		return nil, fmt.Errorf("parsing link: %w", err)
	}

	return link, nil
}

// itemToLink converts a DynamoDB item to a Link model.
func itemToLink(item map[string]types.AttributeValue) (*model.Link, error) {
	link := &model.Link{}

	if v, ok := item["short_code"].(*types.AttributeValueMemberS); ok {
		link.ShortCode = v.Value
		link.ID = v.Value
	}

	if v, ok := item["original_url"].(*types.AttributeValueMemberS); ok {
		link.OriginalURL = v.Value
	}

	if v, ok := item["created_at"].(*types.AttributeValueMemberS); ok {
		t, err := time.Parse(time.RFC3339, v.Value)
		if err != nil {
			return nil, fmt.Errorf("parsing created_at: %w", err)
		}
		link.CreatedAt = t
	}

	if v, ok := item["click_count"].(*types.AttributeValueMemberN); ok {
		var count int64
		_, _ = fmt.Sscanf(v.Value, "%d", &count)
		link.ClickCount = count
	}

	return link, nil
}

// IncrementClickCount atomically increments the click count for a link.
func (r *DynamoLinkRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"short_code": &types.AttributeValueMemberS{Value: shortCode},
		},
		UpdateExpression: aws.String("SET click_count = click_count + :inc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc": &types.AttributeValueMemberN{Value: "1"},
		},
	})

	if err != nil {
		return fmt.Errorf("dynamodb update item: %w", err)
	}

	return nil
}

// Delete removes a link by its short code.
func (r *DynamoLinkRepository) Delete(ctx context.Context, shortCode string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"short_code": &types.AttributeValueMemberS{Value: shortCode},
		},
		ConditionExpression: aws.String("attribute_exists(short_code)"),
	})

	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if ok := errors.As(err, &condErr); ok {
			return repository.ErrNotFound
		}
		return fmt.Errorf("dynamodb delete item: %w", err)
	}

	return nil
}

// DynamoClickRepository implements repository.ClickRepository using DynamoDB.
type DynamoClickRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoClickRepository creates a new DynamoDB-backed click repository.
func NewDynamoClickRepository(tableName string) *DynamoClickRepository {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	return &DynamoClickRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}
}

// Record stores a click event (simplified - just logs for now).
func (r *DynamoClickRepository) Record(ctx context.Context, event *model.ClickEvent) error {
	// For now, we only track click counts (handled by IncrementClickCount).
	// Detailed click events would require a separate table or composite key.
	logger.Debug("click recorded",
		"link_id", event.LinkID,
		"referrer", event.Referrer,
	)
	return nil
}

// GetByLinkID retrieves click events for a link (not implemented yet).
func (r *DynamoClickRepository) GetByLinkID(ctx context.Context, linkID string, limit int) ([]model.ClickEvent, error) {
	// TODO: Implement when we add analytics features
	return []model.ClickEvent{}, nil
}
