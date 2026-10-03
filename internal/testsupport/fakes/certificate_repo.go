package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/certificate"
)

// CertificateRepo is an in-memory certificate.Repository.
type CertificateRepo struct {
	mu          sync.Mutex
	byUserLevel map[string]*certificate.Certificate
	bySerial    map[string]*certificate.Certificate
}

func NewCertificateRepo() *CertificateRepo {
	return &CertificateRepo{
		byUserLevel: make(map[string]*certificate.Certificate),
		bySerial:    make(map[string]*certificate.Certificate),
	}
}

func (f *CertificateRepo) Create(_ context.Context, c *certificate.Certificate) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *c
	f.byUserLevel[c.UserID+"|"+c.LevelID] = &cp
	f.bySerial[c.Serial] = &cp
	return nil
}

func (f *CertificateRepo) FindByUserAndLevel(_ context.Context, userID, levelID string) (certificate.Certificate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byUserLevel[userID+"|"+levelID]
	if !ok {
		return certificate.Certificate{}, certificate.ErrNotFound
	}
	return *c, nil
}

func (f *CertificateRepo) FindBySerial(_ context.Context, serial string) (certificate.Certificate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.bySerial[serial]
	if !ok {
		return certificate.Certificate{}, certificate.ErrNotFound
	}
	return *c, nil
}
