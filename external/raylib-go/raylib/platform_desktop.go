//go:build !android && !drm && !windows
// +build !android,!drm,!windows

package rl

/*
#include "raylib.h"
#include <SDL2/SDL.h>
*/
import "C"
