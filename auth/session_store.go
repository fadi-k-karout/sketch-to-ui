package auth

import (
    "context"
    "fmt"

    "github.com/gofiber/fiber/v2"
    fiberSession "github.com/gofiber/fiber/v2/middleware/session"
)

type fiberSessionStore struct {
    store *fiberSession.Store
}

func NewFiberSessionStore(store *fiberSession.Store) *fiberSessionStore {
    return &fiberSessionStore{store: store}
}

func (s *fiberSessionStore) GetPID(ctx context.Context) (string, error) {
    c, ok := ctx.Value("fiberCtx").(*fiber.Ctx)
    if !ok {
        return "", fmt.Errorf("could not get fiber context")
    }
    sess, err := s.store.Get(c)
    if err != nil {
        return "", err
    }
    raw := sess.Get("ab_pid")
    pid, ok := raw.(string)
    if !ok || pid == "" {
        return "", fmt.Errorf("pid not found or not a string")
    }
    return pid, nil
}

func (s *fiberSessionStore) SetPID(ctx context.Context, pid string) error {
    c, ok := ctx.Value("fiberCtx").(*fiber.Ctx)
    if !ok {
        return fmt.Errorf("could not get fiber context")
    }
    sess, err := s.store.Get(c)
    if err != nil {
        return err
    }
    sess.Set("ab_pid", pid)
    return sess.Save()
}

// Get retrieves any key from the session.
func (s *fiberSessionStore) Get(ctx context.Context, key string) (interface{}, error) {
    c, ok := ctx.Value("fiberCtx").(*fiber.Ctx)
    if !ok {
        return nil, fmt.Errorf("could not get fiber context")
    }
    sess, err := s.store.Get(c)
    if err != nil {
        return nil, err
    }
    return sess.Get(key), nil
}

// Set writes any key/value into the session and persists it.
func (s *fiberSessionStore) Set(ctx context.Context, key string, value interface{}) error {
    c, ok := ctx.Value("fiberCtx").(*fiber.Ctx)
    if !ok {
        return fmt.Errorf("could not get fiber context")
    }
    sess, err := s.store.Get(c)
    if err != nil {
        return err
    }
    sess.Set(key, value)
    return sess.Save()
}

// Delete removes a key from the session and persists the change.
func (s *fiberSessionStore) Delete(ctx context.Context, key string) error {
    c, ok := ctx.Value("fiberCtx").(*fiber.Ctx)
    if !ok {
        return fmt.Errorf("could not get fiber context")
    }
    sess, err := s.store.Get(c)
    if err != nil {
        return err
    }
    sess.Delete(key)
    return sess.Save()
}
