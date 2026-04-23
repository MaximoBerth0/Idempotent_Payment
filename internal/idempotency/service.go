package idempotency

import (
	"context"
	"log/slog"
)

type TxFunc func(ctx context.Context, fn func(ctx context.Context) error) error

type Service struct {
	repo   IdempotencyRepository
	log    *slog.Logger
	withTx TxFunc
}

func NewService(repo IdempotencyRepository, logger *slog.Logger, withTx TxFunc) *Service {
	return &Service{repo: repo, log: logger, withTx: withTx}
}

/*
Execute manages idempotency logic and executes the provided handler atomically.
The handler and the idempotency record save share a single transaction so that
the business operation and the record update either both commit or both roll back.
*/
func (s *Service) Execute(
	ctx context.Context,
	key string,
	requestHash string,
	handler func(ctx context.Context) ([]byte, int, error),
) ([]byte, int, error) {

	record := NewIdempotencyRecord(key, requestHash)

	// CreateIfNotExists is committed immediately (outside the tx) so that
	// concurrent requests see the "in-progress" record right away.
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

	// run the handler and save the result inside a single transaction so that
	// the business write and the idempotency record update are atomic.
	var (
		response   []byte
		httpStatus int
		handlerErr error
	)

	txErr := s.withTx(ctx, func(ctx context.Context) error {
		response, httpStatus, handlerErr = handler(ctx)

		if handlerErr != nil {
			record.MarkFailed(response, httpStatus)
		} else {
			record.MarkCompleted(response, httpStatus)
		}

		return s.repo.Save(ctx, record)
	})

	if txErr != nil {
		s.log.ErrorContext(ctx, "transaction failed", "key", key, "error", txErr)
		return nil, 0, txErr
	}

	if handlerErr != nil {
		s.log.ErrorContext(ctx, "handler execution failed", "key", key, "http_status", httpStatus, "error", handlerErr)
		return response, httpStatus, handlerErr
	}

	s.log.InfoContext(ctx, "idempotency record completed", "key", key, "http_status", httpStatus)
	return response, httpStatus, nil
}
