package piano

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hndada/gosu"
	"github.com/hndada/gosu/draws"
)

type StageDrawer struct {
	Field draws.Sprite
	Hint  draws.Sprite
}

// Todo: might add some effect on StageDrawer
func (d StageDrawer) Draw(screen *ebiten.Image) {
	d.Field.Draw(screen, nil)
	d.Hint.Draw(screen, nil)
}

// Bars are fixed: lane itself moves, all bars move same amount.
type BarDrawer struct {
	Sprite   draws.Sprite
	Cursor   float64
	Farthest *Bar
	Nearest  *Bar
}

func (d *BarDrawer) Update(cursor float64) {
	d.Cursor = cursor
	// When Farthest's prevs are still out of screen due to speed change.
	for d.Farthest.Prev != nil &&
		d.Farthest.Prev.Position-d.Cursor > maxPosition+posMargin {
		d.Farthest = d.Farthest.Prev
	}
	// When Farthest is in screen, next note goes fetched if possible.
	for d.Farthest.Next != nil &&
		d.Farthest.Position-d.Cursor <= maxPosition+posMargin {
		d.Farthest = d.Farthest.Next
	}
	// When Nearest is still in screen due to speed change.
	for d.Nearest.Prev != nil &&
		d.Nearest.Position-d.Cursor > minPosition-posMargin {
		d.Nearest = d.Nearest.Prev
	}
	// When Nearest's next is still out of screen, next note goes fetched.
	for d.Nearest.Next != nil &&
		d.Nearest.Next.Position-d.Cursor <= minPosition-posMargin {
		d.Nearest = d.Nearest.Next
	}
}

func (d BarDrawer) Draw(screen *ebiten.Image) {
	if d.Farthest == nil || d.Nearest == nil {
		return
	}
	for b := d.Farthest; b != d.Nearest.Prev; b = b.Prev {
		sprite := d.Sprite
		pos := b.Position - d.Cursor
		sprite.Move(0, -pos)
		sprite.Draw(screen, nil)
	}
}

// Notes are fixed: lane itself moves, all notes move same amount.
type NoteLaneDrawer struct {
	Sprites  [4]draws.Sprite
	Cursor   float64
	Farthest *Note
	Nearest  *Note
}

// Farthest and Nearest are borders of displaying notes.
// All notes are certainly drawn when drawing from Farthest to Nearest.
func (d *NoteLaneDrawer) Update(cursor float64) {
	d.Cursor = cursor
	if d.Farthest == nil || d.Nearest == nil {
		return
	}
	// When Farthest's prevs are still out of screen due to speed change.
	for d.Farthest.Prev != nil &&
		d.Farthest.Prev.Position-d.Cursor > maxPosition+posMargin {
		d.Farthest = d.Farthest.Prev
	}
	// When Farthest is in screen, next note goes fetched if possible.
	for d.Farthest.Next != nil &&
		d.Farthest.Position-d.Cursor <= maxPosition+posMargin {
		d.Farthest = d.Farthest.Next
	}
	// When Nearest is still in screen due to speed change.
	for d.Nearest.Prev != nil &&
		d.Nearest.Position-d.Cursor > minPosition-posMargin {
		d.Nearest = d.Nearest.Prev
	}
	// When Nearest's next is still out of screen, next note goes fetched.
	for d.Nearest.Next != nil &&
		d.Nearest.Next.Position-d.Cursor <= minPosition-posMargin {
		d.Nearest = d.Nearest.Next
	}
}

// Draw from farthest to nearest to make nearer notes priorly exposed.
func (d NoteLaneDrawer) Draw(screen *ebiten.Image) {
	if d.Farthest == nil || d.Nearest == nil {
		return
	}
	for n := d.Farthest; n != nil && n != d.Nearest.Prev; n = n.Prev {
		if n.Type == Tail {
			d.DrawLongBody(screen, n)
		}
		sprite := d.Sprites[n.Type]
		pos := n.Position - d.Cursor
		sprite.Move(0, -pos)
		op := &ebiten.DrawImageOptions{}
		if n.Marked {
			op.ColorM.ChangeHSV(0, 0.3, 0.3)
		}
		sprite.Draw(screen, op)
	}
}

// DrawLongBody draws scaled, corresponding sub-image of Body sprite.
func (d NoteLaneDrawer) DrawLongBody(screen *ebiten.Image, tail *Note) {
	head := tail.Prev
	body := d.Sprites[Body]
	length := tail.Position - head.Position
	length -= -bodyLoss
	ratio := length / body.H()
	op := &ebiten.DrawImageOptions{}
	if ReverseBody {
		body.SetScale(1, -ratio, ebiten.FilterLinear)
	} else {
		body.SetScale(1, ratio, ebiten.FilterLinear)
	}
	if tail.Marked {
		op.ColorM.ChangeHSV(0, 0.3, 0.3)
	}
	ty := head.Position - d.Cursor
	body.Move(0, -ty)
	body.Draw(screen, op)
}

// // DrawLongBody draws scaled, corresponding sub-image of Body sprite.
// func (d NoteLaneDrawer) DrawLongBody(screen *ebiten.Image, tail *Note) {
// 	head := tail.Prev
// 	body := d.Sprites[Body]
// 	length := tail.Position - head.Position
// 	length -= -bodyLoss
// 	ratio := length / body.H()
// 	trimTail := tail.Position - d.Cursor
// 	if trimTail > maxPosition+posMargin {
// 		trimTail = maxPosition + posMargin
// 	}
// 	trimHead := head.Position - d.Cursor
// 	if trimHead < minPosition-posMargin {
// 		trimHead = minPosition - posMargin
// 	}
// 	fmt.Printf("%.f %.f\n", trimTail, trimHead)
// 	// ground := head.Position - d.Cursor
// 	// propMin := 1 - (trimTail-ground)/length
// 	// propMax := 1 - (trimHead-ground)/length
// 	// propMin := 1 - (tail.Position-head.Position)/length
// 	// propMax := 1 - (head.Position-head.Position)/length
// 	// propMin := 1 - (tail.Position-ground)/length
// 	// propMax := 1 - (head.Position-ground)/length
// 	propMin := (tail.Position - d.Cursor - trimTail) / length
// 	propMax := 1 - (trimHead-head.Position-d.Cursor)/length
// 	subBody := body.SubSprite(0, propMin, 1, propMax)
// 	subBody.SetScale(1, ratio, ebiten.FilterLinear)
// 	op := &ebiten.DrawImageOptions{}
// 	if head.Marked {
// 		op.ColorM.ChangeHSV(0, 0.3, 0.3)
// 	}
// 	if ReverseBody {
// 		subBody.SetScale(1, -ratio, ebiten.FilterLinear)
// 	} else {
// 		subBody.SetScale(1, ratio, ebiten.FilterLinear)
// 	}
// 	ty := (trimHead - head.Position) // trimHead
// 	subBody.Move(0, -ty)
// 	subBody.Draw(screen, op)
// }

// KeyDrawer draws KeyDownSprite at least for 30ms, KeyUpSprite otherwise.
// KeyDrawer uses MinCountdown instead of MaxCountdown.
type KeyDrawer struct {
	MinCountdown   int
	Countdowns     []int
	KeyUpSprites   []draws.Sprite
	KeyDownSprites []draws.Sprite
	lastPressed    []bool
	pressed        []bool
}

func NewKeyDrawer(ups, downs []draws.Sprite) KeyDrawer {
	return KeyDrawer{
		MinCountdown:   gosu.TimeToTick(30),
		Countdowns:     make([]int, len(ups)),
		KeyUpSprites:   ups,
		KeyDownSprites: downs,
	}
}
func (d *KeyDrawer) Update(lastPressed, pressed []bool) {
	d.lastPressed = lastPressed
	d.pressed = pressed
	for k, countdown := range d.Countdowns {
		if countdown > 0 {
			d.Countdowns[k]--
		}
	}
	for k, now := range d.pressed {
		last := d.lastPressed[k]
		if !last && now {
			d.Countdowns[k] = d.MinCountdown
		}
	}
}
func (d KeyDrawer) Draw(screen *ebiten.Image) {
	for k, p := range d.pressed {
		if p || d.Countdowns[k] > 0 {
			d.KeyDownSprites[k].Draw(screen, nil)
		} else {
			d.KeyUpSprites[k].Draw(screen, nil)
		}
	}
}

type ComboDrawer struct {
	draws.BaseDrawer
	Sprites    [10]draws.Sprite
	DigitWidth float64
	DigitGap   float64
	Combo      int
}

// Each number has different width. Number 0's width is used as standard.
func NewComboDrawer(sprites [10]draws.Sprite) ComboDrawer {
	return ComboDrawer{
		BaseDrawer: draws.BaseDrawer{
			MaxCountdown: gosu.TimeToTick(2000),
		},
		Sprites:    sprites,
		DigitWidth: sprites[0].W(),
		DigitGap:   ComboDigitGap,
	}
}
func (d *ComboDrawer) Update(combo int) {
	if d.Countdown > 0 {
		d.Countdown--
	}
	if d.Combo != combo {
		d.Combo = combo
		d.Countdown = d.MaxCountdown
	}
}

// ComboDrawer's Draw draws each number at constant x regardless of their widths.
func (d ComboDrawer) Draw(screen *ebiten.Image) {
	if d.MaxCountdown != 0 && d.Countdown == 0 {
		return
	}
	if d.Combo == 0 {
		return
	}
	vs := make([]int, 0)
	for v := d.Combo; v > 0; v /= 10 {
		vs = append(vs, v%10) // Little endian.
	}

	// Size of the whole image is 0.5w + (n-1)(w+gap) + 0.5w.
	// Since sprites are already at origin, no need to care of two 0.5w.
	w := d.DigitWidth + d.DigitGap
	tx := float64(len(vs)-1) * w / 2
	for _, v := range vs {
		sprite := d.Sprites[v]
		sprite.Move(tx, 0)
		age := d.Age()
		h := sprite.H()
		switch {
		case age < 0.05:
			sprite.Move(0, 0.85*age*h)
		case age >= 0.05 && age < 0.1:
			sprite.Move(0, 0.85*(0.1-age)*h)
		}
		sprite.Draw(screen, nil)
		tx -= w
	}
}

type JudgmentDrawer struct {
	draws.BaseDrawer
	Sprites  []draws.Sprite
	Judgment gosu.Judgment
}

func NewJudgmentDrawer() (d JudgmentDrawer) {
	return JudgmentDrawer{
		BaseDrawer: draws.BaseDrawer{
			MaxCountdown: gosu.TimeToTick(600),
		},
		Sprites: GeneralSkin.JudgmentSprites,
	}
}
func (d *JudgmentDrawer) Update(worst gosu.Judgment) {
	if d.Countdown <= 0 {
		d.Judgment = gosu.Judgment{}
	} else {
		d.Countdown--
	}
	if worst.Window != 0 {
		d.Judgment = worst
		d.Countdown = d.MaxCountdown
	}
}

func (d JudgmentDrawer) Draw(screen *ebiten.Image) {
	if d.Countdown <= 0 {
		return
	}
	var sprite draws.Sprite
	for i, j := range Judgments {
		if j.Window == d.Judgment.Window {
			sprite = d.Sprites[i]
			break
		}
	}
	age := d.Age()
	ratio := 1.0
	switch {
	case age < 0.1:
		ratio = 1.15 * (1 + age)
	case age >= 0.1 && age < 0.2:
		ratio = 1.15 * (1.2 - age)
	case age > 0.9:
		ratio = 1 - 1.15*(age-0.9)
	}
	sprite.SetScale(ratio, ratio, ebiten.FilterLinear)
	sprite.Draw(screen, nil)
}
