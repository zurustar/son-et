package vm

import (
	"fmt"

	"github.com/zurustar/son-et/pkg/graphics"
)

// registerGraphicsBuiltins registers graphics-related built-in functions.
// This includes picture operations, window operations, cast operations,
// text drawing, and shape drawing functions.
func (vm *VM) registerGraphicsBuiltins() {
	// ===== Picture Operations =====

	// LoadPic: Load a picture
	vm.RegisterBuiltinFunction("LoadPic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("LoadPic called but graphics system not initialized", "args", args)
			return -1, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("LoadPic requires filename argument")
		}
		filename, ok := args[0].(string)
		if !ok {
			v.log.Error("LoadPic filename must be string", "got", fmt.Sprintf("%T", args[0]))
			return -1, nil
		}
		picID, err := v.graphicsSystem.LoadPic(filename)
		if err != nil {
			return -1, fmt.Errorf("LoadPic failed")
		}
		v.log.Debug("LoadPic called", "filename", filename, "picID", picID)
		return picID, nil
	})

	// CreatePic: Create a picture
	// Supports three patterns:
	// - CreatePic(srcPicID) - create from existing picture (same size)
	// - CreatePic(width, height) - create with specified size
	// - CreatePic(srcPicID, width, height) - create empty picture with specified size (source is for existence check only)
	vm.RegisterBuiltinFunction("CreatePic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("CreatePic called but graphics system not initialized", "args", args)
			return -1, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("CreatePic requires at least one argument")
		}

		// Check if it's a single argument (create from existing picture)
		if len(args) == 1 {
			srcID, ok := toInt64(args[0])
			if !ok {
				return -1, fmt.Errorf("CreatePic source picture ID must be integer")
			}
			picID, err := v.graphicsSystem.CreatePicFrom(int(srcID))
			if err != nil {
				return -1, fmt.Errorf("CreatePic (from source) failed")
			}
			v.log.Debug("CreatePic called (from source)", "srcID", srcID, "picID", picID)
			return picID, nil
		}

		// Three arguments: srcPicID, width, height
		if len(args) == 3 {
			srcID, sok := toInt64(args[0])
			width, wok := toInt64(args[1])
			height, hok := toInt64(args[2])
			if !sok || !wok || !hok {
				return -1, fmt.Errorf("CreatePic arguments must be integers")
			}
			picID, err := v.graphicsSystem.CreatePicWithSize(int(srcID), int(width), int(height))
			if err != nil {
				return -1, fmt.Errorf("CreatePic (with size) failed: %v", err)
			}
			v.log.Debug("CreatePic called (with size)", "srcID", srcID, "width", width, "height", height, "picID", picID)
			return picID, nil
		}

		// Two arguments: width and height
		width, wok := toInt64(args[0])
		height, hok := toInt64(args[1])
		if !wok || !hok {
			return -1, fmt.Errorf("CreatePic width and height must be integers")
		}
		picID, err := v.graphicsSystem.CreatePic(int(width), int(height))
		if err != nil {
			return -1, fmt.Errorf("CreatePic failed")
		}
		v.log.Debug("CreatePic called", "width", width, "height", height, "picID", picID)
		return picID, nil
	})

	// MovePic: Transfer picture region
	// MovePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y) - mode defaults to 0
	// MovePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, mode)
	// MovePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, mode, speed)
	// MovePic(src_pic, dst_pic) - copy entire picture (simplified form)
	vm.RegisterBuiltinFunction("MovePic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("MovePic called but graphics system not initialized", "args", args)
			return nil, nil
		}

		// Simplified form: MovePic(src_pic, dst_pic) - copy entire picture
		if len(args) == 2 {
			srcID, sok := toInt64(args[0])
			dstID, dok := toInt64(args[1])
			if !sok || !dok {
				return nil, fmt.Errorf("MovePic: picture IDs must be integers")
			}
			// Get source picture dimensions
			srcW := v.graphicsSystem.PicWidth(int(srcID))
			srcH := v.graphicsSystem.PicHeight(int(srcID))
			if err := v.graphicsSystem.MovePic(int(srcID), 0, 0, srcW, srcH, int(dstID), 0, 0, 0); err != nil {
				v.log.Error("MovePic failed", "srcID", srcID, "dstID", dstID, "error", err)
			}
			return nil, nil
		}

		// 6-argument form: MovePic(src, srcX, srcY, width, height, dst) - dstX=0, dstY=0, mode=0
		if len(args) == 6 {
			srcID, _ := toInt64(args[0])
			srcX, _ := toInt64(args[1])
			srcY, _ := toInt64(args[2])
			width, _ := toInt64(args[3])
			height, _ := toInt64(args[4])
			dstID, _ := toInt64(args[5])

			if err := v.graphicsSystem.MovePic(int(srcID), int(srcX), int(srcY), int(width), int(height),
				int(dstID), 0, 0, 0); err != nil {
				v.log.Error("MovePic failed", "error", err)
			}
			return nil, nil
		}

		// Check for invalid argument counts (not 2, 6, and less than 8)
		if len(args) < 8 {
			return nil, fmt.Errorf("MovePic: invalid argument count %d (expected 2, 6, 8, 9, or 10)", len(args))
		}

		srcID, _ := toInt64(args[0])
		srcX, _ := toInt64(args[1])
		srcY, _ := toInt64(args[2])
		width, _ := toInt64(args[3])
		height, _ := toInt64(args[4])
		dstID, _ := toInt64(args[5])
		dstX, _ := toInt64(args[6])
		dstY, _ := toInt64(args[7])

		// mode is optional, defaults to 0 (normal copy)
		var mode int64 = 0
		if len(args) >= 9 {
			mode, _ = toInt64(args[8])
		}

		// Optional speed argument
		speed := 50 // default speed
		if len(args) >= 10 {
			if s, ok := toInt64(args[9]); ok {
				speed = int(s)
			}
		}

		var err error
		if len(args) >= 10 {
			err = v.graphicsSystem.MovePicWithSpeed(int(srcID), int(srcX), int(srcY), int(width), int(height),
				int(dstID), int(dstX), int(dstY), int(mode), speed)
		} else {
			err = v.graphicsSystem.MovePic(int(srcID), int(srcX), int(srcY), int(width), int(height),
				int(dstID), int(dstX), int(dstY), int(mode))
		}

		if err != nil {
			v.log.Error("MovePic failed", "error", err)
		}
		return nil, nil
	})

	// DelPic: Delete a picture
	vm.RegisterBuiltinFunction("DelPic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DelPic called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("DelPic requires picture ID argument")
		}
		picID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("DelPic picture ID must be integer")
			return nil, nil
		}
		if err := v.graphicsSystem.DelPic(int(picID)); err != nil {
			v.log.Error("DelPic failed", "picID", picID, "error", err)
		}
		v.log.Debug("DelPic called", "picID", picID)
		return nil, nil
	})

	// PicWidth: Get picture width
	vm.RegisterBuiltinFunction("PicWidth", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("PicWidth called but graphics system not initialized", "args", args)
			return 0, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("PicWidth requires picture ID argument")
		}
		picID, ok := toInt64(args[0])
		if !ok {
			return 0, fmt.Errorf("PicWidth picture ID must be integer")
		}
		width := v.graphicsSystem.PicWidth(int(picID))
		v.log.Debug("PicWidth called", "picID", picID, "width", width)
		return width, nil
	})

	// PicHeight: Get picture height
	vm.RegisterBuiltinFunction("PicHeight", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("PicHeight called but graphics system not initialized", "args", args)
			return 0, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("PicHeight requires picture ID argument")
		}
		picID, ok := toInt64(args[0])
		if !ok {
			return 0, fmt.Errorf("PicHeight picture ID must be integer")
		}
		height := v.graphicsSystem.PicHeight(int(picID))
		v.log.Debug("PicHeight called", "picID", picID, "height", height)
		return height, nil
	})

	// MoveSPic: Scale and transfer picture
	// MoveSPic(src_pic, src_x, src_y, src_w, src_h, dst_pic, dst_x, dst_y, dst_w, dst_h)
	vm.RegisterBuiltinFunction("MoveSPic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("MoveSPic called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 10 {
			return nil, fmt.Errorf("MoveSPic requires 10 arguments")
		}

		srcID, _ := toInt64(args[0])
		srcX, _ := toInt64(args[1])
		srcY, _ := toInt64(args[2])
		srcW, _ := toInt64(args[3])
		srcH, _ := toInt64(args[4])
		dstID, _ := toInt64(args[5])
		dstX, _ := toInt64(args[6])
		dstY, _ := toInt64(args[7])
		dstW, _ := toInt64(args[8])
		dstH, _ := toInt64(args[9])

		if err := v.graphicsSystem.MoveSPic(int(srcID), int(srcX), int(srcY), int(srcW), int(srcH),
			int(dstID), int(dstX), int(dstY), int(dstW), int(dstH)); err != nil {
			v.log.Error("MoveSPic failed", "error", err)
		}
		v.log.Debug("MoveSPic called", "srcID", srcID, "dstID", dstID)
		return nil, nil
	})

	// TransPic: Transfer with transparency
	// TransPic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, trans_color)
	vm.RegisterBuiltinFunction("TransPic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("TransPic called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 9 {
			return nil, fmt.Errorf("TransPic requires 9 arguments")
		}

		srcID, _ := toInt64(args[0])
		srcX, _ := toInt64(args[1])
		srcY, _ := toInt64(args[2])
		width, _ := toInt64(args[3])
		height, _ := toInt64(args[4])
		dstID, _ := toInt64(args[5])
		dstX, _ := toInt64(args[6])
		dstY, _ := toInt64(args[7])
		transColor, _ := toInt64(args[8])

		if err := v.graphicsSystem.TransPic(int(srcID), int(srcX), int(srcY), int(width), int(height),
			int(dstID), int(dstX), int(dstY), int(transColor)); err != nil {
			v.log.Error("TransPic failed", "error", err)
		}
		v.log.Debug("TransPic called", "srcID", srcID, "dstID", dstID)
		return nil, nil
	})

	// ReversePic: Transfer with horizontal flip
	// ReversePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y)
	vm.RegisterBuiltinFunction("ReversePic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("ReversePic called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 8 {
			return nil, fmt.Errorf("ReversePic requires 8 arguments")
		}

		srcID, _ := toInt64(args[0])
		srcX, _ := toInt64(args[1])
		srcY, _ := toInt64(args[2])
		width, _ := toInt64(args[3])
		height, _ := toInt64(args[4])
		dstID, _ := toInt64(args[5])
		dstX, _ := toInt64(args[6])
		dstY, _ := toInt64(args[7])

		if err := v.graphicsSystem.ReversePic(int(srcID), int(srcX), int(srcY), int(width), int(height),
			int(dstID), int(dstX), int(dstY)); err != nil {
			v.log.Error("ReversePic failed", "error", err)
		}
		v.log.Debug("ReversePic called", "srcID", srcID, "dstID", dstID)
		return nil, nil
	})

	// ===== Window Operations =====

	// OpenWin: Open a window
	vm.RegisterBuiltinFunction("OpenWin", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("OpenWin called but graphics system not initialized", "args", args)
			return -1, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("OpenWin requires at least picture ID argument")
		}
		picID, ok := toInt64(args[0])
		if !ok {
			return -1, fmt.Errorf("OpenWin picture ID must be integer")
		}
		// Pass remaining args as options (will be handled by GraphicsSystem)
		v.log.Debug("OpenWin args", "picID", picID, "opts", args[1:], "optsLen", len(args)-1)
		winID, err := v.graphicsSystem.OpenWin(int(picID), args[1:]...)
		if err != nil {
			return -1, fmt.Errorf("OpenWin failed")
		}
		v.log.Debug("OpenWin called", "picID", picID, "winID", winID)
		return winID, nil
	})

	// CloseWin: Close a window
	vm.RegisterBuiltinFunction("CloseWin", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("CloseWin called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("CloseWin requires window ID argument")
		}
		winID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("CloseWin window ID must be integer")
			return nil, nil
		}
		if err := v.graphicsSystem.CloseWin(int(winID)); err != nil {
			v.log.Error("CloseWin failed", "winID", winID, "error", err)
		}
		v.log.Debug("CloseWin called", "winID", winID)
		return nil, nil
	})

	// MoveWin: Move a window
	vm.RegisterBuiltinFunction("MoveWin", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("MoveWin called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("MoveWin requires at least window ID argument")
		}
		winID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("MoveWin window ID must be integer")
			return nil, nil
		}
		// Pass remaining args as options
		if err := v.graphicsSystem.MoveWin(int(winID), args[1:]...); err != nil {
			v.log.Error("MoveWin failed", "winID", winID, "error", err)
		}
		v.log.Debug("MoveWin called", "winID", winID)
		return nil, nil
	})

	// CloseWinAll: Close all windows
	vm.RegisterBuiltinFunction("CloseWinAll", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("CloseWinAll called but graphics system not initialized")
			return nil, nil
		}
		v.graphicsSystem.CloseWinAll()
		v.log.Debug("CloseWinAll called")
		return nil, nil
	})

	// CapTitle: Set window caption
	// CapTitle(title) - set caption for ALL windows (受け入れ基準 3.1)
	// CapTitle(win_no, title) - set caption for specific window (受け入れ基準 3.3)
	// 存在しないウィンドウIDの場合はエラーを発生させない (受け入れ基準 3.4)
	// 空文字列をタイトルとして受け入れる (受け入れ基準 3.5)
	vm.RegisterBuiltinFunction("CapTitle", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("CapTitle called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("CapTitle requires at least 1 argument")
		}

		if len(args) == 1 {
			// CapTitle(title) - set caption for ALL windows (受け入れ基準 3.1)
			title, _ := args[0].(string)
			v.graphicsSystem.CapTitleAll(title)
			v.log.Debug("CapTitle called (all windows)", "title", title)
		} else {
			// CapTitle(win_no, title) - set caption for specific window (受け入れ基準 3.3)
			winID, _ := toInt64(args[0])
			title, _ := args[1].(string)
			// エラーを無視する (受け入れ基準 3.4: 存在しないウィンドウIDでもエラーを発生させない)
			_ = v.graphicsSystem.CapTitle(int(winID), title)
			v.log.Debug("CapTitle called", "winID", winID, "title", title)
		}
		return nil, nil
	})

	// GetPicNo: Get picture number associated with a window
	vm.RegisterBuiltinFunction("GetPicNo", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("GetPicNo called but graphics system not initialized", "args", args)
			return -1, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("GetPicNo requires window ID argument")
		}
		winID, ok := toInt64(args[0])
		if !ok {
			return -1, fmt.Errorf("GetPicNo window ID must be integer")
		}
		picID, err := v.graphicsSystem.GetPicNo(int(winID))
		if err != nil {
			return -1, fmt.Errorf("GetPicNo failed")
		}
		v.log.Debug("GetPicNo called", "winID", winID, "picID", picID)
		return picID, nil
	})

	// WinInfo: Get virtual desktop information
	// WinInfo(0) - returns width
	// WinInfo(1) - returns height
	vm.RegisterBuiltinFunction("WinInfo", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("WinInfo called but graphics system not initialized", "args", args)
			// Return default values (skelton要件: 1024x768)
			if len(args) >= 1 {
				infoType, _ := toInt64(args[0])
				if infoType == 0 {
					return 1024, nil // default width
				}
				return 768, nil // default height
			}
			return 0, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("WinInfo requires 1 argument")
		}

		infoType, _ := toInt64(args[0])
		var result int
		if infoType == 0 {
			result = v.graphicsSystem.GetVirtualWidth()
		} else {
			result = v.graphicsSystem.GetVirtualHeight()
		}
		v.log.Debug("WinInfo called", "infoType", infoType, "result", result)
		return result, nil
	})

	// ===== Cast Operations =====

	// PutCast: Put a cast on a window
	// PutCast(win_no, pic_no, x, y, src_x, src_y, width, height) - 8 args
	// PutCast(pic_no, base_pic, x, y) - 4 args: simplified form (no transparency)
	// PutCast(pic_no, base_pic, x, y, transparentColor) - 5 args: with transparency
	// PutCast(pic_no, base_pic, x, y, transparentColor, ?, ?, ?, width, height, srcX, srcY) - 12 args: full
	vm.RegisterBuiltinFunction("PutCast", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("PutCast called but graphics system not initialized", "args", args)
			return -1, nil
		}

		// 12 args: PutCast(pic_no, base_pic, x, y, transparentColor, ?, ?, ?, width, height, srcX, srcY)
		if len(args) >= 12 {
			picID, _ := toInt64(args[0])
			basePic, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			transColorInt, _ := toInt64(args[4])
			// args[5-7] = unknown (ignored)
			width, _ := toInt64(args[8])
			height, _ := toInt64(args[9])
			srcX, _ := toInt64(args[10])
			srcY, _ := toInt64(args[11])

			// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
			// basePicが配置先ピクチャー、picIDがソースピクチャー
			transColor := graphics.ColorFromInt(int(transColorInt))
			castID, err := v.graphicsSystem.PutCastWithTransColor(int(picID), int(basePic), int(x), int(y), int(srcX), int(srcY), int(width), int(height), transColor)
			if err != nil {
				v.log.Warn("PutCast failed", "error", err)
				return -1, nil
			}
			v.log.Debug("PutCast called (12 args)", "srcPicID", picID, "dstPicID", basePic, "x", x, "y", y, "transColor", transColorInt, "castID", castID)
			return castID, nil
		}

		// 5 args: PutCast(pic_no, base_pic, x, y, transparentColor)
		if len(args) == 5 {
			picID, _ := toInt64(args[0])
			basePic, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			transColorInt, _ := toInt64(args[4])
			w := v.graphicsSystem.PicWidth(int(picID))
			h := v.graphicsSystem.PicHeight(int(picID))

			// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
			// basePicが配置先ピクチャー、picIDがソースピクチャー
			transColor := graphics.ColorFromInt(int(transColorInt))
			castID, err := v.graphicsSystem.PutCastWithTransColor(int(picID), int(basePic), int(x), int(y), 0, 0, w, h, transColor)
			if err != nil {
				v.log.Warn("PutCast failed", "error", err)
				return -1, nil
			}
			v.log.Debug("PutCast called (5 args)", "srcPicID", picID, "dstPicID", basePic, "transColor", transColorInt, "castID", castID)
			return castID, nil
		}

		// 4 args: PutCast(pic_no, base_pic, x, y) - no transparency
		// スプライトシステムでは、キャストはスプライトとして管理されるため、
		// ピクチャーに焼き付ける必要はない。
		// 注意: _old_implementation2では、キャストを作成した後にbase_picにも画像を焼き付けていたが、
		// スプライトシステムではキャストスプライトが直接描画されるため、この処理は不要。
		if len(args) == 4 {
			picID, _ := toInt64(args[0])
			basePic, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			w := v.graphicsSystem.PicWidth(int(picID))
			h := v.graphicsSystem.PicHeight(int(picID))

			// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
			// basePicが配置先ピクチャー、picIDがソースピクチャー
			castID, err := v.graphicsSystem.PutCast(int(picID), int(basePic), int(x), int(y), 0, 0, w, h)
			if err != nil {
				v.log.Warn("PutCast (4 args) failed", "error", err)
				return -1, nil
			}

			// スプライトシステムでは、キャストはスプライトとして管理されるため、
			// ピクチャーに焼き付ける必要はない（MovePicを呼び出さない）

			v.log.Debug("PutCast called (4 args)", "srcPicID", picID, "dstPicID", basePic, "x", x, "y", y, "castID", castID)
			return castID, nil
		}

		// 8 args: PutCast(src_pic_no, dst_pic_no, x, y, src_x, src_y, width, height)
		// 新API: 第1引数はソースピクチャー、第2引数は配置先ピクチャー
		if len(args) >= 8 {
			srcPicID, _ := toInt64(args[0])
			dstPicID, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			srcX, _ := toInt64(args[4])
			srcY, _ := toInt64(args[5])
			width, _ := toInt64(args[6])
			height, _ := toInt64(args[7])

			castID, err := v.graphicsSystem.PutCast(int(srcPicID), int(dstPicID), int(x), int(y), int(srcX), int(srcY), int(width), int(height))
			if err != nil {
				v.log.Warn("PutCast failed", "error", err)
				return -1, nil
			}
			v.log.Debug("PutCast called (8 args)", "srcPicID", srcPicID, "dstPicID", dstPicID, "castID", castID)
			return castID, nil
		}

		v.log.Error("PutCast: invalid number of arguments", "count", len(args))
		return nil, fmt.Errorf("PutCast requires 4, 5, 8, or 12 arguments, got %d", len(args))
	})

	// MoveCast: Move a cast
	// MoveCast(cast_no, pic_no, x, y, ?, width, height, srcX, srcY) - 9 args (full form)
	// MoveCast(cast_no, x, y) - 3 args: move position only
	// MoveCast(cast_no, x, y, src_x, src_y, width, height) - 7 args: move and change source
	// MoveCast(cast_no, pic_no, x, y) - 4 args: change picture and position
	vm.RegisterBuiltinFunction("MoveCast", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("MoveCast called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 3 {
			return nil, fmt.Errorf("MoveCast requires at least 3 arguments, got %d", len(args))
		}

		castID, _ := toInt64(args[0])

		// 9 args: MoveCast(cast_no, pic_no, x, y, ?, width, height, srcX, srcY)
		// Full form with picture ID and source region
		if len(args) >= 9 {
			picID, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			// args[4] = unknown/transparent color (ignored)
			width, _ := toInt64(args[5])
			height, _ := toInt64(args[6])
			srcX, _ := toInt64(args[7])
			srcY, _ := toInt64(args[8])

			v.log.Debug("MoveCast called (9 args)", "castID", castID, "picID", picID, "x", x, "y", y, "width", width, "height", height, "srcX", srcX, "srcY", srcY)

			opts := []graphics.CastOption{
				graphics.WithCastPicID(int(picID)),
				graphics.WithCastPosition(int(x), int(y)),
				graphics.WithCastSource(int(srcX), int(srcY), int(width), int(height)),
			}
			if err := v.graphicsSystem.MoveCastWithOptions(int(castID), opts...); err != nil {
				v.log.Warn("MoveCast failed", "castID", castID, "error", err)
			}
			return nil, nil
		}

		// 7 args: MoveCast(cast_no, x, y, src_x, src_y, width, height)
		if len(args) == 7 {
			x, _ := toInt64(args[1])
			y, _ := toInt64(args[2])
			srcX, _ := toInt64(args[3])
			srcY, _ := toInt64(args[4])
			width, _ := toInt64(args[5])
			height, _ := toInt64(args[6])

			opts := []graphics.CastOption{
				graphics.WithCastPosition(int(x), int(y)),
				graphics.WithCastSource(int(srcX), int(srcY), int(width), int(height)),
			}
			if err := v.graphicsSystem.MoveCastWithOptions(int(castID), opts...); err != nil {
				v.log.Warn("MoveCast failed", "castID", castID, "error", err)
			}
			return nil, nil
		}

		// 4 args: MoveCast(cast_no, pic_no, x, y)
		if len(args) == 4 {
			picID, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])

			opts := []graphics.CastOption{
				graphics.WithCastPicID(int(picID)),
				graphics.WithCastPosition(int(x), int(y)),
			}
			if err := v.graphicsSystem.MoveCastWithOptions(int(castID), opts...); err != nil {
				v.log.Warn("MoveCast failed", "castID", castID, "error", err)
			}
			return nil, nil
		}

		// 3 args: MoveCast(cast_no, x, y)
		if len(args) == 3 {
			x, _ := toInt64(args[1])
			y, _ := toInt64(args[2])

			opts := []graphics.CastOption{
				graphics.WithCastPosition(int(x), int(y)),
			}
			if err := v.graphicsSystem.MoveCastWithOptions(int(castID), opts...); err != nil {
				v.log.Warn("MoveCast failed", "castID", castID, "error", err)
			}
			return nil, nil
		}

		// Fallback: use old method
		if err := v.graphicsSystem.MoveCast(int(castID), args[1:]...); err != nil {
			v.log.Warn("MoveCast failed", "castID", castID, "error", err)
		}
		return nil, nil
	})

	// DelCast: Delete a cast
	vm.RegisterBuiltinFunction("DelCast", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DelCast called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("DelCast requires cast ID argument")
		}

		castID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("DelCast cast ID must be integer")
			return nil, nil
		}
		if err := v.graphicsSystem.DelCast(int(castID)); err != nil {
			v.log.Error("DelCast failed", "castID", castID, "error", err)
		}
		v.log.Debug("DelCast called", "castID", castID)
		return nil, nil
	})

	// ===== Text Drawing =====

	// TextWrite: Write text to a picture
	// TextWrite(text, pic_no, x, y)
	vm.RegisterBuiltinFunction("TextWrite", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("TextWrite called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 4 {
			return nil, fmt.Errorf("TextWrite requires 4 arguments (text, pic_no, x, y)")
		}

		text, ok := args[0].(string)
		if !ok {
			v.log.Error("TextWrite text must be string", "got", fmt.Sprintf("%T", args[0]))
			return nil, nil
		}
		picID, _ := toInt64(args[1])
		x, _ := toInt64(args[2])
		y, _ := toInt64(args[3])

		if err := v.graphicsSystem.TextWrite(int(picID), int(x), int(y), text); err != nil {
			v.log.Error("TextWrite failed", "error", err)
		}
		v.log.Debug("TextWrite called", "text", text, "picID", picID, "x", x, "y", y)
		return nil, nil
	})

	// SetFont: Set font for text rendering
	// SetFont(size, name, charset, italic, underline, strikeout, weight)
	vm.RegisterBuiltinFunction("SetFont", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("SetFont called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 2 {
			return nil, fmt.Errorf("SetFont requires at least 2 arguments (size, name)")
		}

		size, _ := toInt64(args[0])
		name, _ := args[1].(string)

		// Pass remaining args as options
		if err := v.graphicsSystem.SetFont(name, int(size), args[2:]...); err != nil {
			v.log.Error("SetFont failed", "error", err)
		}
		v.log.Debug("SetFont called", "size", size, "name", name)
		return nil, nil
	})

	// TextColor: Set text color
	// TextColor(r, g, b) or TextColor(color)
	vm.RegisterBuiltinFunction("TextColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("TextColor called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("TextColor requires at least 1 argument")
		}

		var colorInt int
		if len(args) >= 3 {
			// RGB format
			r, _ := toInt64(args[0])
			g, _ := toInt64(args[1])
			b, _ := toInt64(args[2])
			colorInt = int(r)<<16 | int(g)<<8 | int(b)
		} else {
			// Single color value
			colorInt64, _ := toInt64(args[0])
			colorInt = int(colorInt64)
		}

		if err := v.graphicsSystem.SetTextColor(colorInt); err != nil {
			v.log.Error("TextColor failed", "error", err)
		}
		v.log.Debug("TextColor called", "color", fmt.Sprintf("0x%06X", colorInt))
		return nil, nil
	})

	// BgColor: Set background color for text
	// BgColor(r, g, b) or BgColor(color)
	vm.RegisterBuiltinFunction("BgColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("BgColor called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("BgColor requires at least 1 argument")
		}

		var colorInt int
		if len(args) >= 3 {
			// RGB format
			r, _ := toInt64(args[0])
			g, _ := toInt64(args[1])
			b, _ := toInt64(args[2])
			colorInt = int(r)<<16 | int(g)<<8 | int(b)
		} else {
			// Single color value
			colorInt64, _ := toInt64(args[0])
			colorInt = int(colorInt64)
		}

		if err := v.graphicsSystem.SetBgColor(colorInt); err != nil {
			v.log.Error("BgColor failed", "error", err)
		}
		v.log.Debug("BgColor called", "color", fmt.Sprintf("0x%06X", colorInt))
		return nil, nil
	})

	// BackMode: Set background mode for text
	// BackMode(mode) - 0=transparent, 1=opaque
	vm.RegisterBuiltinFunction("BackMode", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("BackMode called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("BackMode requires 1 argument")
		}

		mode, _ := toInt64(args[0])
		if err := v.graphicsSystem.SetBackMode(int(mode)); err != nil {
			v.log.Error("BackMode failed", "error", err)
		}
		v.log.Debug("BackMode called", "mode", mode)
		return nil, nil
	})

	// ===== Shape Drawing =====

	// DrawRect: Draw a rectangle
	// DrawRect(pic_no, x1, y1, x2, y2, fill_mode)
	// DrawRect(pic_no, x1, y1, x2, y2, r, g, b) - with color (fill mode 0)
	vm.RegisterBuiltinFunction("DrawRect", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DrawRect called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 6 {
			return nil, fmt.Errorf("DrawRect requires at least 6 arguments")
		}

		picID, _ := toInt64(args[0])
		x1, _ := toInt64(args[1])
		y1, _ := toInt64(args[2])
		x2, _ := toInt64(args[3])
		y2, _ := toInt64(args[4])
		fillMode, _ := toInt64(args[5])

		// If 8 arguments, treat as DrawRect with color (r, g, b)
		if len(args) >= 8 {
			r, _ := toInt64(args[5])
			g, _ := toInt64(args[6])
			b, _ := toInt64(args[7])
			colorInt := int(r)<<16 | int(g)<<8 | int(b)
			if err := v.graphicsSystem.SetPaintColor(colorInt); err != nil {
				v.log.Error("DrawRect SetPaintColor failed", "error", err)
			}
			fillMode = 0 // outline only when color is specified
		}

		if err := v.graphicsSystem.DrawRect(int(picID), int(x1), int(y1), int(x2), int(y2), int(fillMode)); err != nil {
			v.log.Error("DrawRect failed", "error", err)
		}
		v.log.Debug("DrawRect called", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2, "fillMode", fillMode)
		return nil, nil
	})

	// DrawLine: Draw a line
	// DrawLine(pic_no, x1, y1, x2, y2)
	vm.RegisterBuiltinFunction("DrawLine", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DrawLine called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 5 {
			return nil, fmt.Errorf("DrawLine requires 5 arguments")
		}

		picID, _ := toInt64(args[0])
		x1, _ := toInt64(args[1])
		y1, _ := toInt64(args[2])
		x2, _ := toInt64(args[3])
		y2, _ := toInt64(args[4])

		if err := v.graphicsSystem.DrawLine(int(picID), int(x1), int(y1), int(x2), int(y2)); err != nil {
			v.log.Error("DrawLine failed", "error", err)
		}
		v.log.Debug("DrawLine called", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2)
		return nil, nil
	})

	// FillRect: Fill a rectangle with color
	// FillRect(pic_no, x1, y1, x2, y2, color)
	vm.RegisterBuiltinFunction("FillRect", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("FillRect called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 6 {
			return nil, fmt.Errorf("FillRect requires 6 arguments")
		}

		picID, _ := toInt64(args[0])
		x1, _ := toInt64(args[1])
		y1, _ := toInt64(args[2])
		x2, _ := toInt64(args[3])
		y2, _ := toInt64(args[4])
		colorVal, _ := toInt64(args[5])

		if err := v.graphicsSystem.FillRect(int(picID), int(x1), int(y1), int(x2), int(y2), int(colorVal)); err != nil {
			v.log.Error("FillRect failed", "error", err)
		}
		v.log.Debug("FillRect called", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2)
		return nil, nil
	})

	// DrawCircle: Draw a circle
	// DrawCircle(pic_no, x, y, radius, fill_mode)
	vm.RegisterBuiltinFunction("DrawCircle", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DrawCircle called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 5 {
			return nil, fmt.Errorf("DrawCircle requires 5 arguments")
		}

		picID, _ := toInt64(args[0])
		x, _ := toInt64(args[1])
		y, _ := toInt64(args[2])
		radius, _ := toInt64(args[3])
		fillMode, _ := toInt64(args[4])

		if err := v.graphicsSystem.DrawCircle(int(picID), int(x), int(y), int(radius), int(fillMode)); err != nil {
			v.log.Error("DrawCircle failed", "error", err)
		}
		v.log.Debug("DrawCircle called", "picID", picID, "x", x, "y", y, "radius", radius, "fillMode", fillMode)
		return nil, nil
	})

	// SetLineSize: Set line thickness
	vm.RegisterBuiltinFunction("SetLineSize", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("SetLineSize called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("SetLineSize requires 1 argument")
		}

		size, _ := toInt64(args[0])
		v.graphicsSystem.SetLineSize(int(size))
		v.log.Debug("SetLineSize called", "size", size)
		return nil, nil
	})

	// SetPaintColor: Set paint color
	// SetPaintColor(color) or SetPaintColor(r, g, b)
	vm.RegisterBuiltinFunction("SetPaintColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("SetPaintColor called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("SetPaintColor requires at least 1 argument")
		}

		var colorInt int
		if len(args) >= 3 {
			// RGB format
			r, _ := toInt64(args[0])
			g, _ := toInt64(args[1])
			b, _ := toInt64(args[2])
			colorInt = int(r)<<16 | int(g)<<8 | int(b)
		} else {
			// Single color value
			colorInt64, _ := toInt64(args[0])
			colorInt = int(colorInt64)
		}

		if err := v.graphicsSystem.SetPaintColor(colorInt); err != nil {
			v.log.Error("SetPaintColor failed", "error", err)
		}
		v.log.Debug("SetPaintColor called", "color", fmt.Sprintf("0x%06X", colorInt))
		return nil, nil
	})

	// GetColor: Get pixel color at coordinates
	// GetColor(pic_no, x, y)
	vm.RegisterBuiltinFunction("GetColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("GetColor called but graphics system not initialized", "args", args)
			return 0, nil
		}
		if len(args) < 3 {
			return nil, fmt.Errorf("GetColor requires 3 arguments")
		}

		picID, _ := toInt64(args[0])
		x, _ := toInt64(args[1])
		y, _ := toInt64(args[2])

		colorVal, err := v.graphicsSystem.GetColor(int(picID), int(x), int(y))
		if err != nil {
			return 0, fmt.Errorf("GetColor failed")
		}
		v.log.Debug("GetColor called", "picID", picID, "x", x, "y", y, "color", fmt.Sprintf("0x%06X", colorVal))
		return colorVal, nil
	})

	// SetColor: Set color (alias for SetPaintColor)
	vm.RegisterBuiltinFunction("SetColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("SetColor called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("SetColor requires at least 1 argument")
		}

		var colorInt int
		if len(args) >= 3 {
			// RGB format
			r, _ := toInt64(args[0])
			g, _ := toInt64(args[1])
			b, _ := toInt64(args[2])
			colorInt = int(r)<<16 | int(g)<<8 | int(b)
		} else {
			// Single color value
			colorInt64, _ := toInt64(args[0])
			colorInt = int(colorInt64)
		}

		if err := v.graphicsSystem.SetPaintColor(colorInt); err != nil {
			v.log.Error("SetColor failed", "error", err)
		}
		v.log.Debug("SetColor called", "color", fmt.Sprintf("0x%06X", colorInt))
		return nil, nil
	})
}
