package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type AspectRatioTestCase struct {
	input  string
	width  int
	height int
	Error  error
}

func TestGetVideoAspectRatio(t *testing.T) {
	testCases := map[string]AspectRatioTestCase{
		"valid width and height": {
			input:  "../samples/boots-video-vertical.mp4",
			width:  608,
			height: 1080,
			Error:  nil,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			wantHeight := test.height
			wantWidth := test.width
			haveWidth, haveHeight, err := GetVideoWidthAndHeight(test.input)
			if err != nil {
				t.Error(err)
			}

			outputDiff := cmp.Diff(haveWidth, wantWidth)
			if outputDiff != "" {
				t.Error(outputDiff)
			}
			outputDiff = cmp.Diff(haveHeight, wantHeight)
			if outputDiff != "" {
				t.Error(outputDiff)
			}
		})
	}
}
