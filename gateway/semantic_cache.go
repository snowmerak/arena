package gateway

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

type SemanticCache struct {
	client *redis.Client
	cfg    *EmbeddingConfig
}

func NewSemanticCache(redisUrl string, cfg *EmbeddingConfig) (*SemanticCache, error) {
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)

	ctx := context.Background()
	_ = client.Do(ctx, "FT.CREATE", "idx:cache", "ON", "HASH", "PREFIX", "1", "cache:", "SCHEMA", "vec", "VECTOR", "HNSW", "6", "TYPE", "FLOAT32", "DIM", "1024", "DISTANCE_METRIC", "COSINE").Err()

	return &SemanticCache{
		client: client,
		cfg:    cfg,
	}, nil
}

func float32sToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(f))
	}
	return bytes
}

func (s *SemanticCache) Search(ctx context.Context, vector []float32) (string, error) {
	vecBytes := float32sToBytes(vector)

	// Cosine distance = 1.0 - Cosine Similarity.
	// If similarity >= 0.95, distance <= 0.05
	thresholdDistance := 1.0 - s.cfg.SimilarityThreshold

	query := "*=>[KNN 1 @vec $query_vec AS dist]"

	res, err := s.client.Do(ctx, "FT.SEARCH", "idx:cache", query, "PARAMS", "2", "query_vec", vecBytes, "DIALECT", "2").Result()
	if err != nil {
		return "", err
	}

	results, ok := res.([]interface{})
	if !ok || len(results) < 3 {
		return "", nil 
	}

	fields, ok := results[2].([]interface{})
	if !ok {
		return "", nil
	}

	var dist float64
	var response string

	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			// RediSearch sometimes returns byte arrays
			if b, ok := fields[i].([]byte); ok {
				key = string(b)
			}
		}

		if key == "dist" || key == "__vec_score" {
			if distStr, ok := fields[i+1].(string); ok {
				fmt.Sscanf(distStr, "%f", &dist)
			} else if distB, ok := fields[i+1].([]byte); ok {
				fmt.Sscanf(string(distB), "%f", &dist)
			}
		} else if key == "response" {
			if respStr, ok := fields[i+1].(string); ok {
				response = respStr
			} else if respB, ok := fields[i+1].([]byte); ok {
				response = string(respB)
			}
		}
	}

	if response == "" || dist > thresholdDistance {
		return "", nil 
	}

	return response, nil 
}

func (s *SemanticCache) SearchHash(ctx context.Context, hash string) (string, error) {
	key := fmt.Sprintf("ehash:%s", hash)
	res, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // MISS
	} else if err != nil {
		return "", err
	}
	return res, nil // HIT
}

func (s *SemanticCache) StoreHybrid(ctx context.Context, hash string, vector []float32, response string) error {
	// L1 Exact Hash Storage
	if hash != "" {
		hashKey := fmt.Sprintf("ehash:%s", hash)
		_ = s.client.Set(ctx, hashKey, response, 0).Err()
	}

	// L2 Semantic Vector Storage (only if vector exists)
	if len(vector) > 0 {
		vecBytes := float32sToBytes(vector)
		key := fmt.Sprintf("cache:%d", time.Now().UnixNano())
		return s.client.HSet(ctx, key, "vec", vecBytes, "response", response).Err()
	}
	return nil
}
