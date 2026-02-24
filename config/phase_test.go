package config

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetGlobalPhaseDetectorForTest() {
	globalDetector = nil
	globalDetectorErr = nil
	detectorOnce = sync.Once{}
}

func TestGetGlobalPhaseDetector_InitErrorPersists(t *testing.T) {
	t.Parallel()

	resetGlobalPhaseDetectorForTest()
	originalFactory := phaseDetectorFactory
	t.Cleanup(func() {
		phaseDetectorFactory = originalFactory
		resetGlobalPhaseDetectorForTest()
	})

	calls := 0
	phaseDetectorFactory = func() (*PhaseDetector, error) {
		calls++
		return nil, errors.New("dial failed")
	}

	detector, err := GetGlobalPhaseDetector()
	require.Error(t, err)
	require.Nil(t, detector)

	detector, err = GetGlobalPhaseDetector()
	require.Error(t, err)
	require.Nil(t, detector)
	require.Equal(t, 1, calls)
}

func TestGetGlobalPhaseDetector_NilDetectorReturnsError(t *testing.T) {
	t.Parallel()

	resetGlobalPhaseDetectorForTest()
	originalFactory := phaseDetectorFactory
	t.Cleanup(func() {
		phaseDetectorFactory = originalFactory
		resetGlobalPhaseDetectorForTest()
	})

	phaseDetectorFactory = func() (*PhaseDetector, error) {
		return nil, nil
	}

	detector, err := GetGlobalPhaseDetector()
	require.Error(t, err)
	require.Nil(t, detector)

	phase2Active, err := IsPhase2Active(context.Background())
	require.Error(t, err)
	require.False(t, phase2Active)
}

func TestGetGlobalPhaseDetector_InitializesOnce(t *testing.T) {
	t.Parallel()

	resetGlobalPhaseDetectorForTest()
	originalFactory := phaseDetectorFactory
	t.Cleanup(func() {
		phaseDetectorFactory = originalFactory
		resetGlobalPhaseDetectorForTest()
	})

	expected := &PhaseDetector{}
	calls := 0
	phaseDetectorFactory = func() (*PhaseDetector, error) {
		calls++
		return expected, nil
	}

	detector1, err := GetGlobalPhaseDetector()
	require.NoError(t, err)
	require.Same(t, expected, detector1)

	detector2, err := GetGlobalPhaseDetector()
	require.NoError(t, err)
	require.Same(t, detector1, detector2)
	require.Equal(t, 1, calls)
}
