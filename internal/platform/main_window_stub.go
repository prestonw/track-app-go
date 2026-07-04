//go:build !darwin

package platform

func MainWindowBorderless() {}

func MainWindowMoveBy(_, _ float32) {}

func MainWindowHideAnimated() {}