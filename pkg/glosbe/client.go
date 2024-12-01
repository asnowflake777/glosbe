package glosbe

import (
	"context"
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/url"
	"strings"
)

const (
	endpoint          = "https://glosbe.com"
	translateEndpoint = "https://translate.glosbe.com/"
)

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

type Example struct {
	SrcLangText string
	DstLangText string
}

var ErrNotFound = errors.New("not found")

func (c *Client) Translate(ctx context.Context, req *TranslateRequest) (string, error) {
	uri, err := url.JoinPath(translateEndpoint, req.Src+"-"+req.Dst, req.Text)
	if err != nil {
		return "", err
	}
	resp, err := c.do(ctx, uri)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "Перевода нет(", err
		}
		return "", err
	}
	doc := soup.HTMLParse(resp.String())
	output := doc.Find("app-page-translator-translation-output").Find("div")
	if output.Pointer == nil {
		return "Перевода нет(", nil
	}
	return output.FullText(), nil
}

func (c *Client) Examples(ctx context.Context, req *TranslateRequest) ([]Example, error) {
	uri, err := url.JoinPath(endpoint, req.Src, req.Dst, req.Text)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, uri)
	if err != nil {
		return nil, err
	}
	doc := soup.HTMLParse(resp.String())
	examplesBlock := doc.Find("div", "id", "tmem_first_examples")
	srcExamplesBlck := examplesBlock.FindAll("div", "lang", req.Src)
	dstExamplesBlck := examplesBlock.FindAll("div", "lang", req.Dst)
	if len(srcExamplesBlck) != len(dstExamplesBlck) {
		return nil, fmt.Errorf("unexpected number of examples in src(%d) and dst(%d)",
			len(srcExamplesBlck), len(dstExamplesBlck))
	}
	examples := make([]Example, len(srcExamplesBlck))
	for i := range len(srcExamplesBlck) {
		examples[i] = Example{
			SrcLangText: strings.TrimSpace(srcExamplesBlck[i].FullText()),
			DstLangText: strings.TrimSpace(dstExamplesBlck[i].FullText()),
		}
	}
	return examples, nil
}

func (c *Client) do(ctx context.Context, uri string) (*resty.Response, error) {
	resp, err := c.r.R().
		SetContext(ctx).
		Get(uri)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status())
	}
	return resp, nil
}
