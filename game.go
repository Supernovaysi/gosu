package gosu

import (
	"path/filepath"

	"github.com/hajimehoshi/ebiten"
	"github.com/hndada/gosu/input"
	"github.com/hndada/gosu/mode/mania"
	"github.com/hndada/rg-parser/osugame/osu"
)

// todo: 체크박스 같은거 다시 그리기
// todo: input; ebiten으로 간단히, 나중에 별도 라이브러리.
// todo: PlayIntro, PlayExit
type Game struct {
	Settings
	Scene        Scene
	SceneChanger *SceneChanger
	Skin         Skin
	Input        input.Input
}

// todo: 소리 재생
// Scene이 Game을 control하는 주체
type Scene interface {
	Update(g *Game) error
	Draw(screen *ebiten.Image) // Draws scene to screen
}

func NewGame() (g *Game) {
	g = &Game{}
	g.Settings = LoadSettings()
	g.Scene = g.NewSceneTitle()
	g.SceneChanger = NewSceneChanger()
	return
}
func (g *Game) Update(screen *ebiten.Image) error {
	if !g.SceneChanger.done() {
		return g.SceneChanger.Update(g)
	}
	return g.Scene.Update(g)
}

// 이미지의 method Draw는 input으로 들어온 screen을 그리는 함수
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.SceneChanger.done() {
		g.SceneChanger.Draw(screen)
	}
	g.Scene.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ScreenWidth, g.ScreenHeight
}

// 수정 날짜는 정직하다고 가정
// 플레이 이외의 scene에서, 폴더 변화 감지하면 차트 리로드
// 로드된 차트 데이터는 로컬 db에 별도 저장
func (g *Game) LoadCharts() error {
	for _, d := range dirs {
		for _, f := range files {
			switch filepath.Ext(f.Name()) {
			case ".osu":
				o, err := osu.Parse(path)
				if err != nil {
					panic(err) // todo: log and continue
				}
				switch o.Mode {
				case 3: // todo: osu.ModeMania
					c, err := mania.NewChartFromOsu()
					if err != nil {
						panic(err) // todo: log and continue
					}
					s.Charts = append(s.Charts, c)
				}
			}
		}
	}
	return nil
}
