package celeritas

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"io"
	"strings"
	"time"
)

func (c *Celeritas) ScreenShot(pageURL, testName string, w, h float64) {
	page := rod.New().MustConnect().MustIgnoreCertErrors(true).MustPage(pageURL).MustWaitLoad()

	img, _ := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
		Clip: &proto.PageViewport{
			X:      0,
			Y:      0,
			Width:  w,
			Height: h,
			Scale:  1,
		},
		FromSurface: true,
	})
	fileName := time.Now().Format("2006-01-02-15-04-05.000000")
	_ = utils.OutputFile(fmt.Sprintf("%s/screenshots/%s-%s.png", c.RootPath, testName, fileName), img)
}

func (c *Celeritas) PageHasSelector(pageURL, search string) (bool, error) {
	page := rod.New().MustConnect().MustIgnoreCertErrors(true).MustPage(pageURL).MustWaitLoad()
	has, _, err := page.Has(search)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (c *Celeritas) PageHasText(body io.ReadCloser, search string) (bool, error) {
	resp, err := io.ReadAll(body)
	if err != nil {
		return false, err
	}

	return strings.Contains(string(resp), search), nil
}
