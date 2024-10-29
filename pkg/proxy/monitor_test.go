package proxy

import (
	"context"
	"errors"
	"testing"
	"time"
)

type constantResolver struct {
	value string
}

func (c constantResolver) LookupCNAME(targetCname string) (string, error) {
	return c.value, nil
}

func TestMonitorConstantResolver(t *testing.T) {
	monitorInterval := 100 * time.Millisecond
	m := NewMonitor(constantResolver{"test"}, "test", monitorInterval)
	ctx, cancel := context.WithCancel(context.Background())

	var called bool
	go m.Run(ctx, func() {
		called = true
	})

	time.Sleep(2 * monitorInterval)
	cancel()

	if called {
		t.Error("expected callback not to be called")
	}
}

type changingResolver struct {
	values []string
	errs   []error
	index  int
}

func (c *changingResolver) LookupCNAME(targetCname string) (string, error) {
	if c.index >= len(c.values) {
		c.index = 0
	}
	value := c.values[c.index]
	err := c.errs[c.index]
	c.index++
	return value, err
}

func TestMonitorChangingResolver(t *testing.T) {
	monitorInterval := 100 * time.Millisecond
	m := NewMonitor(&changingResolver{
		values: []string{"test1", "test2"},
		errs:   []error{nil, nil},
		index:  0,
	}, "test", monitorInterval)
	ctx, cancel := context.WithCancel(context.Background())

	var called bool
	go m.Run(ctx, func() {
		called = true
	})

	time.Sleep(2 * monitorInterval)
	cancel()

	if !called {
		t.Error("expected callback to be called")
	}
}

func TestMonitorContinuesOnError(t *testing.T) {
	monitorInterval := 100 * time.Millisecond
	m := NewMonitor(&changingResolver{
		values: []string{"test1", "", "test2"},
		errs:   []error{nil, errors.New("test"), nil},
		index:  0,
	}, "test", monitorInterval)
	ctx, cancel := context.WithCancel(context.Background())

	var called bool
	go m.Run(ctx, func() {
		called = true
	})

	time.Sleep(2 * monitorInterval)
	cancel()

	if !called {
		t.Error("expected callback to be called")
	}
}
