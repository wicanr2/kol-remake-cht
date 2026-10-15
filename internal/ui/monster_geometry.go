package ui

import (
	"fmt"
	"image"
)

// MonsterSheetOrigin 共用原版資訊框的來源尺寸與置中規則。
func MonsterSheetOrigin(size image.Point) (image.Point, error) {
	if size.X <= 0 || size.X%8 != 0 || size.X > 120 || size.Y <= 0 || size.Y > 74 {
		return image.Point{}, fmt.Errorf("MONPICS尺寸無法放入原版資訊框：%d×%d", size.X, size.Y)
	}
	return image.Pt((3+(15-size.X/8)/2)*8, 76-size.Y), nil
}
