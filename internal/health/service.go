package health

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

const (
	manualSource       = "manual"
	manualWeightSource = manualSource
)

type HashGenerator func() (string, error)

type Option func(*service)

type service struct {
	repo          Repository
	now           func() time.Time
	hashGenerator HashGenerator
}

func WithClock(now func() time.Time) Option {
	return func(s *service) {
		s.now = now
	}
}

func WithHashGenerator(fn HashGenerator) Option {
	return func(s *service) {
		s.hashGenerator = fn
	}
}

func NewService(repo Repository, opts ...Option) Service {
	s := &service{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
		hashGenerator: defaultHashGenerator,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func defaultHashGenerator() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
