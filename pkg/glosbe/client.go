package glosbe

import (
	"context"
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/go-resty/resty/v2"
	"net/http"
	"strings"
)

const GlosbeEndpoint = "https://glosbe.com"

type Client struct {
	r *resty.Client
}

func NewClient(r *resty.Client) *Client {
	return &Client{r: r}
}

type TranslateRequest struct {
	Src  string
	Dst  string
	Text string
}

type TranslateResponse struct {
	Examples []Example
}

type Example struct {
	SrcLangText string
	DstLangText string
}

var ErrNotFound = errors.New("not found")

func (c *Client) Translate(ctx context.Context, req *TranslateRequest) (*TranslateResponse, error) {
	resp, err := c.r.R().
		SetContext(ctx).
		Get(fmt.Sprintf("%s/%s/%s/%s", GlosbeEndpoint, req.Src, req.Dst, req.Text))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status())
	}
	doc := soup.HTMLParse(resp.String())
	examplesBlock := doc.Find("div", "id", "tmem_first_examples")
	srcExamplesBlck := examplesBlock.FindAll("div", "lang", req.Src)
	dstExamplesBlck := examplesBlock.FindAll("div", "lang", req.Dst)
	if len(srcExamplesBlck) != len(dstExamplesBlck) {
		return nil, fmt.Errorf("unexpected number of examples in src(%d) and dst(%d)",
			len(srcExamplesBlck), len(dstExamplesBlck))
	}
	res := &TranslateResponse{
		Examples: make([]Example, len(srcExamplesBlck)),
	}
	for i := range len(srcExamplesBlck) {
		res.Examples[i] = Example{
			SrcLangText: strings.TrimSpace(srcExamplesBlck[i].FullText()),
			DstLangText: strings.TrimSpace(dstExamplesBlck[i].FullText()),
		}
	}
	return res, nil
}
