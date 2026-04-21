package idempotency

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo IdempotencyRepository
	log  *slog.Logger
	pool *pgxpool.Pool
}

func NewService(repo IdempotencyRepository, logger *slog.Logger, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, log: logger, pool: pool}
}

/*
this is a higher-order function, Execute() manages idempotency logic and executes the provided handler function.
it is executed atomically
*/

func (s *Service) Execute(
	ctx context.Context,
	key string,
	requestHash string,
	handler func(ctx context.Context) ([]byte, int, error),
) ([]byte, int, error) {

	record := NewIdempotencyRecord(key, requestHash)

	existing, created, err := s.repo.CreateIfNotExists(ctx, record)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to create idempotency record", "key", key, "error", err)
		return nil, 0, err
	}

	if !created {
		s.log.DebugContext(ctx, "idempotency record already exists", "key", key, "status", existing.Status)

		if err := existing.ValidateHash(requestHash); err != nil {
			s.log.WarnContext(ctx, "idempotency hash mismatch", "key", key, "error", err)
			return nil, 0, err
		}

		switch existing.Status {
		case StatusCompleted, StatusFailed:
			s.log.InfoContext(ctx, "returning cached response", "key", key, "status", existing.Status, "http_status", existing.HTTPStatus)
			return existing.Response, existing.HTTPStatus, nil

		case StatusInProgress:
			s.log.WarnContext(ctx, "request already in progress", "key", key)
			return nil, 409, ErrRequestInProgress
		}
	}

	response, httpStatus, handlerErr := handler(ctx)

	if handlerErr != nil {
		s.log.ErrorContext(ctx, "handler execution failed", "key", key, "http_status", httpStatus, "error", handlerErr)
		record.MarkFailed(response, httpStatus)
		if saveErr := s.repo.Save(ctx, record); saveErr != nil {
			s.log.ErrorContext(ctx, "failed to save failed idempotency record", "key", key, "error", saveErr)
		}
		return response, httpStatus, handlerErr
	}

	record.MarkCompleted(response, httpStatus)
	if err := s.repo.Save(ctx, record); err != nil {
		s.log.ErrorContext(ctx, "failed to save completed idempotency record", "key", key, "error", err)
		return nil, 0, err
	}

	s.log.InfoContext(ctx, "idempotency record completed", "key", key, "http_status", httpStatus)
	return response, httpStatus, nil
}
