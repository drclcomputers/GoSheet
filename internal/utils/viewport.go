// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// viewport.go provides the definition and functions for managing the table viewport

package utils

type Viewport struct {
    TopRow    int32
    LeftCol   int32
    ViewRows  int32
    ViewCols  int32
}


func (vp *Viewport) ToAbsolute(visualRow, visualCol int32) (int32, int32) {
    return vp.TopRow + visualRow - 1, vp.LeftCol + visualCol - 1
}

func (vp *Viewport) ToRelative(absRow, absCol int32) (int32, int32) {
    return absRow - vp.TopRow + 1, absCol - vp.LeftCol + 1
}

func (vp *Viewport) IsVisible(absRow, absCol int32) bool {
    return absRow >= vp.TopRow && absRow < vp.TopRow+vp.ViewRows &&
           absCol >= vp.LeftCol && absCol < vp.LeftCol+vp.ViewCols
}

