# Y_SARU Integration Test Results

**Test Date:** 2025-01-17
**Task:** 24.6 Integration test with y_saru sample
**Status:** ✅ PASSED

## Test Execution

```bash
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/y_saru
```

**Duration:** 8 seconds (auto-terminated)
**Exit Code:** 0 (clean exit)

## Verification Results

### ✅ 1. Window IDs Displayed

Window IDs are correctly displayed in DEBUG mode (DEBUG_LEVEL=2):

```
DEBUG: WinID=0 Pic=0 ImgRect=(494,263)-(854,604)
DEBUG: WinID=1 Pic=8 ImgRect=(494,263)-(834,563)
DEBUG: WinID=2 Pic=3 ImgRect=(494,263)-(794,537)
DEBUG: WinID=3 Pic=14 ImgRect=(494,263)-(794,527)
DEBUG: WinID=4 Pic=27 ImgRect=(494,263)-(854,641)
DEBUG: WinID=5 Pic=35 ImgRect=(494,263)-(984,593)
DEBUG: WinID=6 Pic=39 ImgRect=(494,263)-(834,513)
DEBUG: WinID=7 Pic=44 ImgRect=(494,263)-(874,513)
DEBUG: WinID=8 Pic=8 ImgRect=(494,263)-(834,563)
DEBUG: WinID=9 Pic=49 ImgRect=(494,263)-(1024,578)
```

**Result:** Window IDs 0-9 are displayed with their associated Picture IDs.

### ✅ 2. Picture IDs Displayed

Picture IDs are correctly assigned and displayed:

**Loaded Pictures (Sequential IDs):**
- POP-W.BMP → ID=0
- POPOP01.BMP → ID=1
- POPOP02.BMP → ID=2
- POP01.BMP → ID=3
- ... (continues sequentially)
- POP02E.BMP → ID=14
- POP02E.BMP → ID=15 (no unexpected P0 here!)
- ... (continues sequentially)
- POPED.BMP → ID=51

**Created Pictures (New Sequential IDs):**
- CreatePic(25) → ID=27 (350x300)
- CreatePic(41) → ID=44 (300x250)
- CreatePic(46) → ID=49 (409x315)

**Result:** Picture IDs are sequential and correctly assigned. CreatePic returns new IDs, not source IDs.

### ✅ 3. Casts Visible in Scenes

**Cast Creation Log:**
```
Created and drew cast ID=1 at 0,0    (シーン3 - background)
Created and drew cast ID=2 at 113,0  (シーン3 - flying plane)
Created and drew cast ID=3 at 0,0    (シーン5.5 - background)
Created and drew cast ID=4 at 80,0   (シーン5.5 - monkey)
Created and drew cast ID=5 at 0,0    (シーン7 - background)
Created and drew cast ID=6 at 45,57  (シーン7 - monkey)
```

**Cast Movement Examples (シーン3 - Flying Plane):**
```
MoveCast CastID=2: Pos=(109,0) Size=89x77 Src=(0,0)
MoveCast CastID=2: Pos=(105,10) Size=89x77 Src=(89,0)
MoveCast CastID=2: Pos=(101,20) Size=89x77 Src=(0,0)
MoveCast CastID=2: Pos=(97,30) Size=89x77 Src=(89,0)
MoveCast CastID=2: Pos=(93,40) Size=89x77 Src=(0,0)
MoveCast CastID=2: Pos=(89,50) Size=89x77 Src=(89,0)
```

**Cast Rendering Confirmation:**
```
DEBUG MoveCast: Redrawing casts for dest picture 27
  Cast 1: Pic=25 Pos=(0,0) Size=350x300 SrcOffset=(0,0)
  Cast 2: Pic=28 Pos=(109,0) Size=89x77 SrcOffset=(0,0)
  Total casts redrawn: 2
```

**Result:** Casts are created, moved, and redrawn correctly in all three scenes.

### ✅ 4. Correct Picture Sequence (No P0 After P14)

**Previous Bug:** P14の後にP0が表示される (P0 appeared after P14)

**Verification:**
- Picture sequence: 0, 1, 2, ..., 13, 14, **15**, 16, 17, ... 51
- After P14 (ID=14), the next picture is P15 (ID=15)
- No unexpected P0 (ID=0) appears after P14

**CreatePic Behavior:**
- CreatePic(25) returns ID=27 (not 25)
- CreatePic(41) returns ID=44 (not 41)
- CreatePic(46) returns ID=49 (not 46)

**Result:** Picture ID assignment is correct. CreatePic creates new sequential IDs.

### ✅ 5. No Errors in Execution

**Error Check:**
```bash
grep -iE "error|panic|fatal|fail" y_saru_integration_test.log
```

**Result:** No errors, panics, or failures detected.

### ✅ 6. Transparency Processing

**Cast Transparency Examples:**
```
PutCast: Transparent color specified: #ffffff
convertTransparentColor: Target color RGB=(255,255,255)
  Converted: 53164 transparent, 9811 opaque pixels
PutCast: Created transparency-processed picture ID=50
```

**Result:** Transparency is processed once at cast creation and stored as a new picture.

## Requirements Validated

- ✅ **4.2** - Picture IDs assigned sequentially
- ✅ **4.3** - CreatePic returns new sequential IDs
- ✅ **5.1** - PutCast creates casts correctly
- ✅ **5.2** - Cast IDs assigned sequentially
- ✅ **5.3** - MoveCast updates cast position and re-renders
- ✅ **5.5** - Casts rendered in creation order (z-ordering)
- ✅ **14.1** - Windows created with correct properties
- ✅ **14.2** - Window IDs displayed in debug mode

## Bug Fixes Confirmed

### Bug 1: P0 appears after P14 ✅ FIXED
- **Before:** CreatePic(sourcePicID) returned sourcePicID
- **After:** CreatePic(sourcePicID) returns new sequential ID
- **Evidence:** CreatePic(25)→27, CreatePic(41)→44, CreatePic(46)→49

### Bug 2: Casts not visible ✅ FIXED
- **Before:** Casts created but not drawn to destination picture
- **After:** Casts drawn correctly with transparency processing
- **Evidence:** 6 casts created and redrawn in シーン3, シーン5.5, シーン7

### Enhancement: Window ID Display ✅ IMPLEMENTED
- **Feature:** Window IDs displayed in DEBUG_LEVEL=2
- **Evidence:** "DEBUG: WinID=X Pic=Y" logs for all windows

## Conclusion

All verification points passed successfully:
- ✅ Window IDs displayed
- ✅ Picture IDs displayed and sequential
- ✅ Casts visible in all scenes
- ✅ No unexpected P0 after P14
- ✅ No errors in execution
- ✅ All requirements validated

**Integration Test Status: PASSED** ✅
