# Cast Transparency Implementation Fix

## Date: 2026-01-15

## Problem

Cast transparency was not working correctly in sample scenarios. Sprites with white backgrounds were rendering as opaque white rectangles instead of being transparent.

## Root Cause

The implementation had a critical design flaw where transparency processing was being applied at the wrong time:

1. **Initial incorrect approach**: `MoveCast` was applying transparency processing on every draw call
2. **Performance issue**: `convertTransparentColor()` loops through all pixels - doing this on every draw is extremely inefficient
3. **Design violation**: The code was processing transparency multiple times instead of once

## Correct Design Principle

**Separation of Concerns:**
- **Picture (Pic)**: Raw image data loaded from files. NEVER modified after loading.
- **Cast**: A sprite that references a Picture, with optional transparency processing.

**Transparency Processing Strategy:**

### At Cast Creation Time (PutCast):
1. When `PutCast` is called with a transparent color parameter:
   - Call `convertTransparentColor()` ONCE to create a transparency-processed image
   - Store this processed image as a NEW Picture with a new Picture ID
   - The Cast references this processed Picture ID (NOT the original)
   - Draw the processed image immediately to the destination

### At Draw Time (MoveCast):
1. Simply draw the transparency-processed Picture that's already stored in the Cast
2. NO additional transparency processing needed
3. Ebitengine's native alpha blending handles the transparency automatically

## Example Flow

```
LoadPic("sprite.bmp") → Picture ID 17 (original image with white background)
PutCast(17, dest, x, y, 0xffffff, ...) → 
  1. convertTransparentColor(Picture 17, white) → new image with transparency
  2. Store as Picture ID 28 (transparency-processed)
  3. Create Cast ID 2 with Picture=28
  4. Draw Picture 28 to destination
MoveCast(2, ...) →
  1. Cast 2 references Picture 28 (already processed)
  2. Draw Picture 28 to destination (no processing needed)
```

## Performance Considerations

- Transparency processing is expensive (pixel-by-pixel loop)
- MUST be done only ONCE at Cast creation time
- NEVER process transparency on every draw call
- The processed image is reused for all subsequent draws

## Code Changes

### File: `pkg/engine/engine.go`

**In `PutCast()` function (lines ~2330-2450):**
- Added transparency processing at Cast creation time
- Creates new Picture ID for processed image
- Stores processed Picture ID in Cast
- Draws the Cast immediately

**In `MoveCast()` function (lines ~2600-2650):**
- REMOVED incorrect transparency processing at draw time
- Added comment explaining why transparency is NOT processed here
- Cast.Picture field already references the transparency-processed image

## Documentation Updates

### File: `.kiro/specs/core-engine/design.md`

Added comprehensive documentation in the "Cast Manager" section (lines ~300-380) explaining:
- The design principle of Picture vs Cast separation
- Transparency processing strategy (once at creation, not at draw time)
- Performance considerations
- Example flow
- Common mistakes to avoid

## Common Mistakes to Avoid

- ❌ Processing transparency on every draw call (performance issue)
- ❌ Modifying the original Picture (violates separation of concerns)
- ❌ Storing transparency info in Cast but processing at draw time (inefficient)
- ✅ Process transparency ONCE at Cast creation, store as new Picture
- ✅ Keep original Picture unchanged
- ✅ Draw the processed Picture directly (no additional processing)

## Testing

Verified with sample scenarios:
- Transpiled successfully
- Built without errors
- Sprites now render with correct transparency (white backgrounds are transparent)

## Lessons Learned

1. **Design before implementation**: Should have clarified the Picture vs Cast separation earlier
2. **Performance matters**: Pixel-by-pixel operations must be minimized
3. **Documentation is critical**: Complex design decisions must be documented to prevent regression
4. **Test with real samples**: Sample scenarios revealed issues that unit tests might have missed

## Related Files

- `pkg/engine/engine.go` - Implementation
- `.kiro/specs/core-engine/design.md` - Design documentation
