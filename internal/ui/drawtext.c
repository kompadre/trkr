#include <stdint.h>
#include "raylib.h"

#define BATCH_SIZE 1024
#define TEXT_BUFFER_SIZE 8192

typedef struct {
    uint8_t r, g, b, a;
} BatchColor;

typedef struct {
    char text_pool[TEXT_BUFFER_SIZE];
    uint32_t text_offsets[BATCH_SIZE]; // Index into text_pool for each string
    float xs[BATCH_SIZE];
    float ys[BATCH_SIZE];
    float sizes[BATCH_SIZE];
    BatchColor colors[BATCH_SIZE];
    int count;
    int pool_next;
} TextBatch;

// Inline function to reset the batch indicators instantly without zeroing memory
void ResetBatch(TextBatch* b) {
    b->count = 0;
    b->pool_next = 0;
}

int AppendToBatch(TextBatch* b, const char* str, int len, float x, float y, float size, uint8_t r, uint8_t g, uint8_t b_val, uint8_t a) {
    if (b->count >= BATCH_SIZE || (b->pool_next + len + 1) >= TEXT_BUFFER_SIZE) {
        return 0; // Batch full
    }

    // Set text offset pointing to current pool pointer position
    b->text_offsets[b->count] = b->pool_next;

    // Copy characters manually into the flat array
    for (int i = 0; i < len; i++) {
        b->text_pool[b->pool_next++] = str[i];
    }
    b->text_pool[b->pool_next++] = '\0'; // Append crucial null terminator safely on C side

    b->xs[b->count] = x;
    b->ys[b->count] = y;
    b->sizes[b->count] = size;
    b->colors[b->count] = (BatchColor){r, g, b_val, a};

    b->count++;
    return 1;
}

void FlushTextBatch(TextBatch* b, Font font) {
    if (b->count == 0) return;

    for (int i = 0; i < b->count; i++) {
        uint32_t offset = b->text_offsets[i];
        const char* c_str = &b->text_pool[offset];

        Vector2 pos = { b->xs[i], b->ys[i] };
        Color col = { b->colors[i].r, b->colors[i].g, b->colors[i].b, b->colors[i].a };

        // Calling native C Raylib directly! Zero Go string involvement.
        DrawTextEx(font, c_str, pos, b->sizes[i], 0.0f, col);
        // DrawText(c_str, b->xs[i], b->ys[i], b->sizes[i], col);
    }

     ResetBatch(b);
}


