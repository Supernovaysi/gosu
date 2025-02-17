package piano

import (
	"fmt"
	"image/color"
	"math"

	_ "image/jpeg"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hndada/gosu"
	"github.com/hndada/gosu/draws"
)

type NoteKind int

const (
	One NoteKind = iota
	Two
	Mid
	Tip = Mid
)

var NoteKindsMap = map[int][]NoteKind{
	0:  {},
	1:  {Mid},
	2:  {One, One},
	3:  {One, Mid, One},
	4:  {One, Two, Two, One},
	5:  {One, Two, Mid, Two, One},
	6:  {One, Two, One, One, Two, One},
	7:  {One, Two, One, Mid, One, Two, One},
	8:  {Tip, One, Two, One, One, Two, One, Tip},
	9:  {Tip, One, Two, One, Mid, One, Two, One, Tip},
	10: {Tip, One, Two, One, Mid, Mid, One, Two, One, Tip},
}

// LeftScratch and RightScratch are bits for indicating scratch mode.
// For example, when key count is 40 = 32 + 8, it is 8-key with left scratch.
const (
	LeftScratch  = 32
	RightScratch = 64
	ScratchMask  = ^(LeftScratch | RightScratch)
)

func init() { // I'm proud of the following code.
	for k := 2; k <= 8; k++ {
		NoteKindsMap[k|LeftScratch] = append([]NoteKind{Tip}, NoteKindsMap[k-1]...)
		NoteKindsMap[k|RightScratch] = append(NoteKindsMap[k-1], Tip)
	}
}

var GeneralSkin struct { // Singleton
	ComboSprites    [10]draws.Sprite
	JudgmentSprites []draws.Sprite
}

// Todo: should each skin has own skin settings?
// Order of fields in Skin is focused on readability.
type Skin struct {
	ScoreSprites [10]draws.Sprite
	SignSprites  [3]draws.Sprite

	ComboSprites    [10]draws.Sprite
	JudgmentSprites []draws.Sprite // Todo: slice -> array?

	KeyUpSprites   []draws.Sprite
	KeyDownSprites []draws.Sprite
	NoteSprites    []draws.Sprite
	HeadSprites    []draws.Sprite
	TailSprites    []draws.Sprite
	BodySprites    []draws.Sprite
	// BodySprites    [][]draws.Sprite // Binary-building method

	FieldSprite draws.Sprite
	HintSprite  draws.Sprite
	BarSprite   draws.Sprite // Seperator of each bar (aka measure)
}

var Skins = make(map[int]Skin)

func LoadSkin() {
	// Sprites that are independent of key count.
	for i := 0; i < 10; i++ {
		s := draws.NewSprite(fmt.Sprintf("skin/combo/%d.png", i))
		s.SetScale(ComboScale)
		s.SetPosition(FieldPosition, ComboPosition, draws.OriginCenterMiddle)
		GeneralSkin.ComboSprites[i] = s
	}
	GeneralSkin.JudgmentSprites = make([]draws.Sprite, 5)
	for i, name := range []string{"kool", "cool", "good", "bad", "miss"} {
		s := draws.NewSprite(fmt.Sprintf("skin/piano/judgment/%s.png", name))
		s.SetScale(JudgmentScale)
		s.SetPosition(FieldPosition, JudgmentPosition, draws.OriginCenterMiddle)
		GeneralSkin.JudgmentSprites[i] = s
	}

	// Following sprites are dependent of key count.
	// Todo: animated sprite support. Starting with [4][]*ebiten.Image will help.
	var (
		keyUpImage   *ebiten.Image
		keyDownImage *ebiten.Image
		hintImage    *ebiten.Image
		noteImages   [4]*ebiten.Image
		headImages   [4]*ebiten.Image
		tailImages   [4]*ebiten.Image
		// bodyImages   [4]image.Image // binary-building method
		bodyImages [4]*ebiten.Image
	)
	keyUpImage = draws.NewImage("skin/piano/key/up.png") // Todo: combine with declaration?
	keyDownImage = draws.NewImage("skin/piano/key/down.png")
	hintImage = draws.NewImage("skin/piano/hint.png")
	// Todo: 4th note image. 1st note with custom color settings.
	for i, kind := range []int{1, 2, 3, 3} {
		noteImages[i] = draws.NewImage(fmt.Sprintf("skin/piano/note/note/%d.png", kind))
		headImages[i] = draws.NewImage(fmt.Sprintf("skin/piano/note/head/%d.png", kind))
		tailImages[i] = draws.NewImage(fmt.Sprintf("skin/piano/note/tail/%d.png", kind))
		bodyImages[i] = draws.NewImage(fmt.Sprintf("skin/piano/note/body/%d.png", kind))
		// bodyImages[i] = draws.NewImageSrc(fmt.Sprintf("skin/piano/note/body/%d.png", kind))
	}

	// Todo: Key count 1, 2, 3 and with scratch
	for keyCount := 4; keyCount <= 10; keyCount++ {
		noteKinds := NoteKindsMap[keyCount]
		noteWidths := NoteWidthsMap[keyCount]
		skin := Skin{
			ScoreSprites: gosu.ScoreSprites,
			SignSprites:  gosu.SignSprites,

			ComboSprites:    GeneralSkin.ComboSprites,
			JudgmentSprites: GeneralSkin.JudgmentSprites[:],

			KeyUpSprites:   make([]draws.Sprite, keyCount&ScratchMask),
			KeyDownSprites: make([]draws.Sprite, keyCount&ScratchMask),
			NoteSprites:    make([]draws.Sprite, keyCount&ScratchMask),
			HeadSprites:    make([]draws.Sprite, keyCount&ScratchMask),
			TailSprites:    make([]draws.Sprite, keyCount&ScratchMask),
			BodySprites:    make([]draws.Sprite, keyCount&ScratchMask),
			// BodySprites:    make([][]draws.Sprite, keyCount&ScratchMask),
		}
		// KeyUp and KeyDown are drawn below Hint, which bottom is along with HitPosition.
		// Each w should be integer, since it is a width of independent sprite.
		// Todo: Scratch should be excluded to width sum.
		var wsum float64
		for _, kind := range noteKinds {
			wsum += math.Ceil(noteWidths[kind])
		}
		x := FieldPosition - wsum/2
		for k, kind := range noteKinds {
			w := math.Ceil(noteWidths[kind])
			{
				s := draws.NewSpriteFromImage(keyUpImage)
				scaleW := w / s.W()
				scaleH := (screenSizeY - HitPosition) / s.H()
				s.SetScaleXY(scaleW, scaleH, ebiten.FilterLinear)
				s.SetPosition(x, HitPosition, draws.OriginLeftTop)
				skin.KeyUpSprites[k] = s
			}
			{
				s := draws.NewSpriteFromImage(keyDownImage)
				scaleW := w / s.W()
				scaleH := (screenSizeY - HitPosition) / s.H()
				s.SetScaleXY(scaleW, scaleH, ebiten.FilterLinear)
				s.SetPosition(x, HitPosition, draws.OriginLeftTop)
				skin.KeyDownSprites[k] = s
			}
			{
				s := draws.NewSpriteFromImage(noteImages[kind])
				scaleW := w / s.W()
				scaleH := NoteHeigth / s.H()
				s.SetScaleXY(scaleW, scaleH, ebiten.FilterLinear)
				s.SetPosition(x, HitPosition, draws.OriginLeftBottom)
				skin.NoteSprites[k] = s
			}
			{
				s := draws.NewSpriteFromImage(headImages[kind])
				scaleW := w / s.W()
				scaleH := NoteHeigth / s.H()
				s.SetScaleXY(scaleW, scaleH, ebiten.FilterLinear)
				s.SetPosition(x, HitPosition, draws.OriginLeftBottom)
				skin.HeadSprites[k] = s
			}
			{
				s := draws.NewSpriteFromImage(tailImages[kind])
				scaleW := w / s.W()
				scaleH := NoteHeigth / s.H()
				s.SetScaleXY(scaleW, scaleH, ebiten.FilterLinear)
				s.SetPosition(x, HitPosition, draws.OriginLeftBottom)
				skin.TailSprites[k] = s
			}
			{
				s := draws.NewSpriteFromImage(bodyImages[kind])
				s.SetScale(w / s.W())
				s.SetPosition(x, HitPosition, draws.OriginLeftBottom) // Todo: need a test // -NoteHeigth
				skin.BodySprites[k] = s
			}
			x += w
		}
		{
			src := ebiten.NewImage(int(wsum), screenSizeY)
			src.Fill(color.NRGBA{0, 0, 0, uint8(255 * FieldDarkness)})
			s := draws.NewSpriteFromImage(src)
			s.SetPosition(FieldPosition, 0, draws.OriginCenterTop)
			skin.FieldSprite = s
		}
		{
			s := draws.NewSpriteFromImage(hintImage)
			scaleW := wsum / s.W()
			scaleH := HintHeight / s.H()
			s.SetScaleXY(scaleW, scaleH, ebiten.FilterLinear)
			s.SetPosition(FieldPosition, HitPosition-HintHeight, draws.OriginCenterTop)
			skin.HintSprite = s
		}
		{
			src := ebiten.NewImage(int(wsum), 1)
			src.Fill(color.NRGBA{255, 255, 255, 255}) // White
			s := draws.NewSpriteFromImage(src)
			s.SetPosition(FieldPosition, HitPosition, draws.OriginCenterBottom)
			skin.BarSprite = s
		}
		Skins[keyCount] = skin
	}
}

// // Draw max length of long note body sprite in advance.
// src := bodyImages[kind]
// scale := float64(w) / float64(src.Bounds().Dx())
// h := int(scale * float64(src.Bounds().Dy()))
// dst := image.NewRGBA(image.Rect(0, 0, w, screenSizeY))
// switch BodyStyle {
// case BodyStyleStretch:
// 	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
// case BodyStyleAttach:
// 	for rect := image.Rect(0, 0, w, h); rect.Min.Y < dst.Bounds().Dy(); {
// 		draw.BiLinear.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
// 		rect.Min.Y += h
// 		rect.Max.Y += h
// 	}
// }
// for pow := 0; pow < int(math.Log2(screenSizeY))+1; pow++ {
// 	h := 1 << pow
// 	rect := image.Rect(0, 0, int(w), h)
// 	s := draws.NewSpriteFromImage(dst.SubImage(rect))
// 	skin.BodySprites[k] = append(skin.BodySprites[k], draws.Sprite{
// 		I: ebiten.NewImageFromImage(dst.SubImage(rect)),
// 		W: float64(w),
// 		H: float64(h),
// 		X: float64(x),
// 		// BodySprites's y value is not fixed.
// 	})
// }

// This is for test.
// {
// 	w := int(noteWidths[kind])
// 	scale := float64(w) / float64(src.Bounds().Dx())
// 	h := int(scale * float64(src.Bounds().Dy()))
// 	dst := image.NewRGBA(image.Rect(0, 0, int(w), screenSizeY))
// 	for rect := image.Rect(0, 0, w, h); rect.Min.Y < dst.Bounds().Dy(); {
// 		draw.BiLinear.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
// 		rect.Min.Y += h
// 		rect.Max.Y += h
// 	}
// 	s.BodySpritesTest[k] = draws.Sprite{
// 		I: ebiten.NewImageFromImage(dst),
// 		W: float64(dst.Bounds().Dx()), // noteWidths[kind]
// 		H: float64(dst.Bounds().Dy()), // screenSizeY
// 	}
// 	s.BodySpritesTest[k].X = float64(x)
// 	// BodySprites's y value is not fixed.
// }
