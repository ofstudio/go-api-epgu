package apipgu

import (
	"fmt"
	"mime/multipart"
	"net/textproto"
)

type multipartFunc func(w *multipart.Writer) error

func multipartBodyPrepare(w *multipart.Writer, fns ...multipartFunc) error {
	var err error
	for _, fn := range fns {
		if err = fn(w); err != nil {
			return fmt.Errorf("%w: %w", ErrMultipartBodyPrepare, err)
		}
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("%w: %w", ErrMultipartBodyPrepare, err)
	}
	return nil
}

func withMeta(meta OrderMeta) multipartFunc {
	return func(w *multipart.Writer) error {
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
	}
}

func withOrderId(id int) multipartFunc {
	return func(w *multipart.Writer) error {
		return w.WriteField("orderId", fmt.Sprintf("%d", id))
	}
}

func withFile(filename string, data []byte) multipartFunc {
	return func(w *multipart.Writer) error {
		fw, err := w.CreateFormFile("file", filename)
		if err != nil {
			return err
		}
		if _, err = fw.Write(data); err != nil {
			return err
		}
		return nil
	}
}

func withChunkNum(current, total int) multipartFunc {
	return func(w *multipart.Writer) error {
		if err := w.WriteField("chunk", fmt.Sprintf("%d", current)); err != nil {
			return err
		}
		return w.WriteField("chunks", fmt.Sprintf("%d", total))
	}
}
