//go:build !sceneverify

package main

func validateSceneVerification() error       { return nil }
func sceneVerificationSnapshot() any         { return nil }
func prepareSceneVerification(g *game) error { return nil }
