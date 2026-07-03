package ui

/*
#cgo CFLAGS: -I${SRCDIR}/../../external/raylib-go/raylib
#include <stdint.h>
#include "raylib.h"

// Tell CGO what layouts to expect from drawtext.c
typedef struct { uint8_t r, g, b, a; } BatchColor;
typedef struct {
    char text_pool[8192];
    uint32_t text_offsets[1024];
    float xs[1024];
    float ys[1024];
    float sizes[1024];
    BatchColor colors[1024];
    int count;
    int pool_next;
} TextBatch;

void ResetBatch(TextBatch* b);
int AppendToBatch(TextBatch* b, const char* str, int len, float x, float y, float size, uint8_t r, uint8_t g, uint8_t b_val, uint8_t a);
void FlushTextBatch(TextBatch* b, Font font);
*/
import "C"
import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"unsafe"
)

type TextBatcher struct {
	cBatch C.TextBatch
}

func (tb *TextBatcher) Reset() {
	C.ResetBatch(&tb.cBatch)
}

func (tb *TextBatcher) Add(text string, x, y, size float32, color rl.Color) {
	if len(text) == 0 {
		return
	}

	// Zero allocations: read pointer to Go's string backing array directly
	strPtr := (*C.char)(unsafe.Pointer(unsafe.StringData(text)))
	strLen := C.int(len(text))

	C.AppendToBatch(&tb.cBatch, strPtr, strLen,
		C.float(x), C.float(y), C.float(size),
		C.uint8_t(color.R), C.uint8_t(color.G), C.uint8_t(color.B), C.uint8_t(color.A),
	)
}

func (tb *TextBatcher) Flush(font rl.Font) {
	if tb.cBatch.count == 0 {
		return
	}

	// Safely alias the raylib-go structural representation down to raw C memory layouts
	cFont := *(*C.Font)(unsafe.Pointer(&font))

	// Cross the CGO bridge exactly ONCE for the entire pattern matrix
	C.FlushTextBatch(&tb.cBatch, cFont)
}
