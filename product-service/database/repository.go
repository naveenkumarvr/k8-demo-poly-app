package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Product represents a product in the catalog
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductRepository defines the interface for product data operations
// This interface enables easy mocking for testing
type ProductRepository interface {
	GetAllProducts(ctx context.Context) ([]Product, error)
	GetProductByID(ctx context.Context, id int) (*Product, error)
	GetProductsByCategory(ctx context.Context, category string) ([]Product, error)
	CreateProduct(ctx context.Context, product *Product) error
}

// PostgresProductRepository implements ProductRepository using PostgreSQL
type PostgresProductRepository struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

// NewProductRepository creates a new PostgreSQL product repository
func NewProductRepository(client *Client) ProductRepository {
	return &PostgresProductRepository{
		pool:   client.Pool(),
		tracer: otel.Tracer("product-service"),
	}
}

// GetAllProducts retrieves all products from the database
func (r *PostgresProductRepository) GetAllProducts(ctx context.Context) ([]Product, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetAllProducts")
	defer span.End()

	query := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at
		FROM products
		ORDER BY category, name
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "products"),
	)

	startTime := time.Now()
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Stock,
			&p.Category,
			&p.ImageURL,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	duration := time.Since(startTime)
	span.SetAttributes(
		attribute.Int("db.result.count", len(products)),
		attribute.Int64("db.query.duration_ms", duration.Milliseconds()),
	)

	return products, nil
}

// GetProductByID retrieves a single product by its ID
func (r *PostgresProductRepository) GetProductByID(ctx context.Context, id int) (*Product, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetProductByID")
	defer span.End()

	query := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "products"),
		attribute.Int("product.id", id),
	)

	startTime := time.Now()
	var p Product
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.Stock,
		&p.Category,
		&p.ImageURL,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	duration := time.Since(startTime)
	span.SetAttributes(
		attribute.Int64("db.query.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get product by ID %d: %w", id, err)
	}

	return &p, nil
}

// GetProductsByCategory retrieves all products in a specific category
func (r *PostgresProductRepository) GetProductsByCategory(ctx context.Context, category string) ([]Product, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetProductsByCategory")
	defer span.End()

	query := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at
		FROM products
		WHERE category = $1
		ORDER BY name
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "products"),
		attribute.String("product.category", category),
	)

	startTime := time.Now()
	rows, err := r.pool.Query(ctx, query, category)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query products by category: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Stock,
			&p.Category,
			&p.ImageURL,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	duration := time.Since(startTime)
	span.SetAttributes(
		attribute.Int("db.result.count", len(products)),
		attribute.Int64("db.query.duration_ms", duration.Milliseconds()),
	)

	return products, nil
}

// CreateProduct inserts a new product into the database
func (r *PostgresProductRepository) CreateProduct(ctx context.Context, product *Product) error {
	ctx, span := r.tracer.Start(ctx, "repository.CreateProduct")
	defer span.End()

	query := `
		INSERT INTO products (name, description, price, stock, category, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "products"),
		attribute.String("product.name", product.Name),
		attribute.String("product.category", product.Category),
	)

	startTime := time.Now()
	err := r.pool.QueryRow(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.Category,
		product.ImageURL,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	duration := time.Since(startTime)
	span.SetAttributes(
		attribute.Int64("db.query.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create product: %w", err)
	}

	span.SetAttributes(attribute.Int("product.id", product.ID))
	return nil
}
