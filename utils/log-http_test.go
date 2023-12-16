package utils

import "testing"

func Test_sanitizeMultipartFile(t *testing.T) {
	tests := []struct {
		name string
		dump string
		want string
	}{
		{name: "no multipart", dump: noMultipartDump, want: noMultipartWant},
		{name: "multipart binary", dump: multipartBinaryDump, want: multipartBinaryWant},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeMultipartFile(tt.dump); got != tt.want {
				t.Errorf("sanitizeMultipartFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	noMultipartDump = "POST / HTTP/1.1\r\nHost: localhost:8080\r\nContent-Type: application/json\r\n\r\n{\"foo\":\"bar\"}"
	noMultipartWant = noMultipartDump

	multipartBinaryDump = "POST / HTTP/1.1\r\nHost: localhost:8080\r\nContent-Type: multipart/form-data; boundary=--------------------------1234567890\r\n\r\n----------------------------1234567890\r\nContent-Disposition: form-data; name=\"foo\"\r\n\r\nbar\r\n----------------------------1234567890\r\nContent-Disposition: form-data; name=\"file\"; filename=\"file.txt\"\r\nContent-Type: application/octet-stream\r\n\r\nThis is file content\r\n----------------------------1234567890--\r\nContent-Disposition: form-data; name=\"baz\"\r\n\r\nqux\r\n----------------------------1234567890--\r\n"
	multipartBinaryWant = "POST / HTTP/1.1\r\nHost: localhost:8080\r\nContent-Type: multipart/form-data; boundary=--------------------------1234567890\r\n\r\n----------------------------1234567890\r\nContent-Disposition: form-data; name=\"foo\"\r\n\r\nbar\r\n----------------------------1234567890\r\nContent-Disposition: form-data; name=\"file\"; filename=\"file.txt\"\r\nContent-Type: application/octet-stream\r\n\r\n[ 20 bytes of binary data... ]\r\n----------------------------1234567890--\r\nContent-Disposition: form-data; name=\"baz\"\r\n\r\nqux\r\n----------------------------1234567890--\r\n"
)
