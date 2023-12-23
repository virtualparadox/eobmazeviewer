package main

import (
	dat2 "EOB1MazeViewer/formats"
	"EOB1MazeViewer/renderer"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/virtualparadox/xbrscaler"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	zoom         = 4
	screenWidth  = 176 * zoom
	screenHeight = 120 * zoom

	frameWidth  = 32
	frameHeight = 32
)

type Game struct {
	xbrscaler                   *xbrscaler.Xbr
	mazeRenderer                *renderer.MazeRenderer
	x, y, direction             int
	prevX, prevY, prevDirection int
	mazeView                    *ebiten.Image
}

func (g *Game) Update() error {

	// Move
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.moveInMaze(g.direction)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.moveInMaze((g.direction + 2) & 0x03)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.moveInMaze((g.direction - 1) & 0x03)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.moveInMaze((g.direction + 1) & 0x03)
	}

	// Turn
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.direction = (g.direction + 1) & 0x03
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.direction = (g.direction - 1) & 0x03
	}

	if g.needUpdate() {
		g.updateMazeView()
	}

	return nil
}

func (g *Game) updateMazeView() {
	renderedImage, _ := g.mazeRenderer.RenderMaze(g.x, g.y, g.direction)
	palettedImage := BytesToPalettedImage(renderedImage, 176, 120, g.mazeRenderer.Palette.GetPalette())
	rgbaImage := ConvertPalettedToRGBA(palettedImage, true)
	arrayImage := ConvertRGBAtoUint32Array(rgbaImage)
	scaledImage, scaledWidth, scaledHeight := g.xbrscaler.Xbr4x(arrayImage, 176, 120, true, true)

	g.mazeView = ConvertUint32ArrayToEbitenImage(scaledImage, scaledWidth, scaledHeight)

	g.prevX = g.x
	g.prevY = g.y
	g.prevDirection = g.direction
}

func (g *Game) needUpdate() bool {
	return g.x != g.prevX || g.y != g.prevY || g.direction != g.prevDirection
}

func (g *Game) moveInMaze(direction int) {
	g.prevX = g.x
	g.prevY = g.y
	g.prevDirection = g.direction

	switch direction {
	case 0:
		g.y--
	case 1:
		g.x++
	case 2:
		g.y++
	case 3:
		g.x--
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
	op.GeoM.Translate(screenWidth/2, screenHeight/2)

	if g.mazeView != nil {
		screen.DrawImage(g.mazeView, &ebiten.DrawImageOptions{})
	}
	ebitenutil.DebugPrint(screen, "X="+strconv.Itoa(g.x)+" Y="+strconv.Itoa(g.y)+" Direction="+strconv.Itoa(g.direction))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	args := os.Args

	if len(args) != 3 {
		fmt.Printf("Usage: maze-viewer EOB1DATA_DIR LEVEL\neg: maze-viewer /home/joe/EOB1 8")
		os.Exit(1)
	}

	dataFiles := loadDataFiles(args)
	mazeRenderer := initMazeRenderer(args, dataFiles)
	xbrScaler := xbrscaler.NewXbrScaler(false)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("EOB1 - Maze Viewer")

	if err := ebiten.RunGame(&Game{
		xbrscaler:    xbrScaler,
		mazeRenderer: mazeRenderer,
		x:            10,
		y:            15,
		direction:    0}); err != nil {
		log.Fatal(err)
	}

}

func initMazeRenderer(args []string, dataFiles map[string]*[]byte) *renderer.MazeRenderer {
	infName := "LEVEL" + args[2] + ".INF"
	_, ok := dataFiles[infName]
	if !ok {
		log.Fatalf("Cannot find file %s", infName)
	}

	inf, _ := dat2.NewInfFromByteArray(dataFiles[infName])

	mazName := "LEVEL" + args[2] + ".MAZ"
	maz, _ := dat2.NewMazFromByteArray(dataFiles[mazName])
	vcn, _ := dat2.NewVCNFromByteArray(dataFiles[strings.ToUpper(inf.VmpVcnName)+".VCN"])
	vmp, _ := dat2.NewVMPFromByteArray(dataFiles[strings.ToUpper(inf.VmpVcnName)+".VMP"])
	pal, _ := dat2.NewPALFromByteArray(dataFiles[strings.ToUpper(inf.PaletteName)+".PAL"])

	decorationCPSNames := inf.GetDecorationCPSNames()
	dat, _ := dat2.NewDATFromByteArray(dataFiles[strings.ToUpper(inf.VmpVcnName)+".DAT"])
	decorationContainer := renderer.BuildDecorationContainer(dat, dataFiles, decorationCPSNames)

	mazeRenderer := renderer.NewMazeRenderer(inf, maz, vcn, vmp, pal, decorationContainer)
	return mazeRenderer
}

func loadDataFiles(args []string) map[string]*[]byte {
	dataFiles, err := UnPak(args[1])
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%d files loaded into memory.", len(dataFiles))
	return dataFiles
}
