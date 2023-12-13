package apipgu

import (
	"fmt"
	"mime/multipart"
	"net/textproto"
)

type multipartFunc func(w *multipart.Writer) error

type multipartBuilder struct {
	w   *multipart.Writer
	fns []multipartFunc
}

func newMultipartBuilder(w *multipart.Writer) *multipartBuilder {
	return &multipartBuilder{w: w}
}

func (b *multipartBuilder) withMeta(meta OrderMeta) *multipartBuilder {
	b.fns = append(b.fns, func(w *multipart.Writer) error {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="meta"`)
		h.Set("Content-Type", "application/json")
		fw, err := w.CreatePart(h)
		if err != nil {
			return err
		}
		if _, err = fw.Write(meta.JSON()); err != nil {
			return err
		}
		return nil
	})
	return b
}

func (b *multipartBuilder) withOrderId(id int) *multipartBuilder {
	b.fns = append(b.fns, func(w *multipart.Writer) error {
		return w.WriteField("orderId", fmt.Sprintf("%d", id))
	})
	return b
}

func (b *multipartBuilder) withFile(filename string, data []byte) *multipartBuilder {
	b.fns = append(b.fns, func(w *multipart.Writer) error {
		fw, err := w.CreateFormFile("file", filename)
		if err != nil {
			return err
		}
		if _, err = fw.Write(data); err != nil {
			return err
		}
		return nil
	})
	return b
}

func (b *multipartBuilder) withChunkNum(current, total int) *multipartBuilder {
	b.fns = append(b.fns, func(w *multipart.Writer) error {
		if err := w.WriteField("chunk", fmt.Sprintf("%d", current)); err != nil {
			return err
		}
		return w.WriteField("chunks", fmt.Sprintf("%d", total))
	})
	return b
}

func (b *multipartBuilder) build() error {
	var err error
	for _, fn := range b.fns {
		if err = fn(b.w); err != nil {
			return fmt.Errorf("%w: %w", ErrMultipartBody, err)
		}
	}
	if err = b.w.Close(); err != nil {
		return fmt.Errorf("%w: %w", ErrMultipartBody, err)
	}
	return nil
}
