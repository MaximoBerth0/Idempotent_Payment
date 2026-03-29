package idempotency

import "context"

type Service struct {
	repo IdempotencyRepository
}

func NewService(repo IdempotencyRepository) *Service {
	return &Service{repo: repo}
}

// this is a higher-order function, Execute() manages idempotency logic and executes the provided handler function.
func (s *Service) Execute(
	ctx context.Context,
	key string,
	requestHash string,
	handler func(ctx context.Context) ([]byte, int, error),
) ([]byte, int, error) {

	record := NewIdempotencyRecord(key, requestHash)

	existing, created, err := s.repo.CreateIfNotExists(ctx, record)
	if err != nil {
		return nil, 0, err
	}

	if !created {

		if err := existing.ValidateHash(requestHash); err != nil {
			return nil, 0, err
		}

		switch existing.Status {
		case StatusCompleted, StatusFailed:
			return existing.Response, existing.HTTPStatus, nil

		case StatusInProgress:
			return nil, 409, ErrRequestInProgress
		}
	}

	response, httpStatus, handlerErr := handler(ctx)

	if handlerErr != nil {
		record.MarkFailed(response, httpStatus)
		_ = s.repo.Save(ctx, record)
		return response, httpStatus, handlerErr
	}

	record.MarkCompleted(response, httpStatus)
	if err := s.repo.Save(ctx, record); err != nil {
		return nil, 0, err
	}

	return response, httpStatus, nil
}
