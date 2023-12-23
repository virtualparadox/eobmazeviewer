package renderer

import (
	"EOB1MazeViewer/formats"
	"strings"
)

type DecorationContainer struct {
	decorationData *formats.DecorationData
	cpsFileData    map[string]*[]byte
}

func BuildDecorationContainer(decorationData *formats.DecorationData, files map[string]*[]byte, cpsFilenames []string) *DecorationContainer {
	cpsFileData := make(map[string]*[]byte)
	for _, cpsFilename := range cpsFilenames {
		if cpsFilename == "" {
			continue
		}

		fn := strings.ToUpper(cpsFilename) + ".CPS"
		cpsRawData, _ := formats.NewCPSFromByteArray(files[fn])
		cpsFileData[cpsFilename] = cpsRawData.GetRawData()
	}

	return &DecorationContainer{
		decorationData: decorationData,
		cpsFileData:    cpsFileData,
	}
}

func (c DecorationContainer) GetDecoration(id int) formats.Decoration {
	return c.decorationData.Decorations[id]
}

func (c DecorationContainer) GetCPSRectangle(index byte) formats.DecorationRectangle {
	storedRectangle := c.decorationData.Rectangles[index]
	return formats.DecorationRectangle{
		X: storedRectangle.X,
		Y: storedRectangle.Y,
		W: storedRectangle.W,
		H: storedRectangle.H,
	}
}

func (c DecorationContainer) GetDecorationBitmapByName(name string) *[]byte {
	return c.cpsFileData[name]
}
