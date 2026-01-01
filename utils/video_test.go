package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type WidthAndHegihtTestCase struct {
	input  string
	output MetaData
	Error  error
}
type MetaData struct {
	width  int
	height int
}

func TestGetVideoWidthAndHeight(t *testing.T) {
	testCases := map[string]WidthAndHegihtTestCase{
		"valid width and height": {
			input: "../samples/boots-video-vertical.mp4",
			output: MetaData{
				width:  608,
				height: 1080,
			},
			Error: nil,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			wantHeight := test.output.height
			wantWidth := test.output.width
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

type AspectRatioTestCase struct {
	input  MetaData
	output string
}

func TestGetVideoAspectRatio(t *testing.T) {
	testCases := map[string]AspectRatioTestCase{
		"protrait video": {
			input: MetaData{
				width:  608,
				height: 1080,
			},
			output: "9:16",
		},
		"landscape video": {
			input: MetaData{
				width:  1080,
				height: 608,
			},
			output: "16:9",
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			want := test.output
			have := GetVideoAspectRatio(test.input.width, test.input.height)

			outputDiff := cmp.Diff(have, want)
			if outputDiff != "" {
				t.Error(outputDiff)
			}
		})
	}
}
